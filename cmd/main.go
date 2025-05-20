package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/noorbala7418/grafana-oncall-notifier/internal/models"
	"github.com/noorbala7418/grafana-oncall-notifier/pkg/config"
	"github.com/noorbala7418/grafana-oncall-notifier/pkg/email"
	"github.com/sirupsen/logrus"
)

func initLog(logLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	switch logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	cfgPath, cfgErr := config.ParseFlags()
	if cfgErr != nil {
		logrus.Fatal(cfgErr)
	}
	cfg, cfgErr := config.NewConfig(cfgPath)
	if cfgErr != nil {
		logrus.Fatal(cfgErr)
	}

	initLog(cfg.LogLevel)

	logrus.Info("oncall-notifier")
	logrus.Info("log level is: ", cfg.LogLevel)
	logrus.Info("start app")

	run(cfg)
	logrus.Info("end app.")
}

func run(cfg *models.Config) {
	// get correct time
	now := time.Now()
	loc := time.UTC
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 20, 30, 0, 0, loc).Format("2006-01-02T15:04")
	endDate := time.Date(now.Year(), now.Month(), now.Day()+1, 20, 30, 0, 0, loc).Format("2006-01-02T15:04")
	logrus.Info("function run: start date is: ", startDate, ", end date is: ", endDate)

	// get shifts
	shifts, shiftErr := getSchedules(&cfg.Grafana)
	if shiftErr != nil {
		logrus.Error("function run: error in get list of shifts. err: ", shiftErr)
		os.Exit(1)
	}
	logrus.Debug("function run: we have ", shifts.Count, " shifts.")

	for _, shift := range shifts.Shifts {
		logrus.Debug("function run: get oncalls for shift: ", shift.Name)
		// get oncalls from each shifts
		oncalls, oncallsErr := getOncallShifts(&cfg.Grafana, &shift, startDate, endDate)
		if oncallsErr != nil {
			logrus.Error("function run: error in get list of oncalls in shift: ", shift.Name, ". err: ", oncallsErr)
			continue
		}
		logrus.Debug("function run: got oncalls for shift: ", shift.Name, ". going to send notification.")

		// send notification
		sendNotifErr := sendNotification(&cfg.EmailCredential, &shift, oncalls)
		if sendNotifErr != nil {
			logrus.Error("function run: error in send notification for oncalls in shift: ", shift.Name, ". err: ", sendNotifErr)
			continue
		}
		logrus.Info("function run: send notification done for shift: ", shift.Name)
	}
}

func getSchedules(grafana *models.GrafanaOnCallCredentials) (*models.OnCallShifts, error) {
	grafanaUrl := grafana.Url + "/api/v1/schedules"
	logrus.Debug("function getSchedules: grafana api url is: ", grafanaUrl)
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, reqErr := http.NewRequest("GET", grafanaUrl, nil)
	if reqErr != nil {
		logrus.Error("function getSchedules: Error in create request. err: ", reqErr)
		return &models.OnCallShifts{}, fmt.Errorf("function getSchedules: Error in create request. err: %w", reqErr)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", grafana.AdminToken)
	logrus.Debug("function getSchedules: request sent")

	resp, err := client.Do(req)
	if err != nil {
		logrus.Error("function getSchedules: Error in get schedules. err: ", err)
		return &models.OnCallShifts{}, fmt.Errorf("function getSchedules: Error in get schedules. err: %w", err)
	}
	logrus.Debug("function getSchedules: got response")

	if resp != nil {
		body, respErr := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		if respErr != nil {
			logrus.Error("function getSchedules: Error in read body. err: ", respErr)
			return &models.OnCallShifts{}, fmt.Errorf("function getSchedules: Error in read body. err: %w", respErr)
		}
		logrus.Debug("function getSchedules: body is ok")

		var shedules models.OnCallShifts
		if jsonParseErr := json.Unmarshal(body, &shedules); jsonParseErr != nil {
			logrus.Error("function getSchedules: Error in parse json body. err: ", jsonParseErr)
			return &models.OnCallShifts{}, fmt.Errorf("function getSchedules: Error in parse json body. err: %w", jsonParseErr)
		}
		logrus.Debug("function getSchedules: json body parsed.")
		return &shedules, nil
	}
	return &models.OnCallShifts{}, nil
}

func getOncallShifts(grafana *models.GrafanaOnCallCredentials, shedule *models.Shift, startDate string, endDate string) (*models.OnCalls, error) {
	grafanaUrl := grafana.Url + "/api/v1/schedules/" + shedule.ShiftID + "/final_shifts?start_date=" + startDate + "&end_date=" + endDate
	logrus.Debug("function getOncallShifts: grafana oncalls url is: ", grafanaUrl)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, reqErr := http.NewRequest("GET", grafanaUrl, nil)
	if reqErr != nil {
		logrus.Error("function getOncallShifts: Error in create request. err: ", reqErr)
		return &models.OnCalls{}, fmt.Errorf("function getOncallShifts: Error in create request. shedule name is: %s, err: %w", shedule.Name, reqErr)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", grafana.AdminToken)

	resp, err := client.Do(req)
	if err != nil {
		logrus.Error("function getOncallShifts: Error in get shifts. shedule name is: ", shedule.Name, " err: ", err)
		return &models.OnCalls{}, fmt.Errorf("function getOncallShifts: Error in get shifts. shedule name is: %s, err: %w", shedule.Name, err)
	}
	logrus.Debug("function getOncallShifts: request sent.")

	if resp != nil {
		body, respErr := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		if respErr != nil {
			logrus.Error("function getOncallShifts: Error in read body. shedule name is: ", shedule.Name, " err:", respErr)
			return &models.OnCalls{}, fmt.Errorf("function getOncallShifts: Error in read body. shedule name is: %s, err: %w", shedule.Name, respErr)
		}
		logrus.Debug("function getOncallShifts: body is readable.")

		var oncalls models.OnCalls
		if jsonParseErr := json.Unmarshal(body, &oncalls); jsonParseErr != nil {
			logrus.Error("function getOncallShifts: Error in parse json body. shedule name is: ", shedule.Name, " err: ", jsonParseErr)
			return &models.OnCalls{}, fmt.Errorf("function getOncallShifts: Error in parse json body. shedule name is: %s, err: %w", shedule.Name, jsonParseErr)
		}
		logrus.Debug("function getOncallShifts: json body parsed successfully.")

		return &oncalls, nil
	}
	return &models.OnCalls{}, nil
}

func sendNotification(emailCredentials *models.EmailCredential, shift *models.Shift, oncalls *models.OnCalls) error {
	// making names unique
	var receivers []string
	for _, oncall := range oncalls.OnCall {
		receivers = append(receivers, oncall.Username)
	}
	receivers = uniqueNonEmptyElementsOf(receivers)
	logrus.Debug("function sendNotification: unique receviers is: ", receivers)
	logrus.Debug("function sendNotification: length of unique receviers is: ", len(receivers))

	for i := range receivers {
		oncallsList := ""
		timePattern := "2006-01-02T15:04:05Z"
		logrus.Debug("function sendNotification: prepare reminder message for user ", receivers[i], " in shift: ", shift.Name)

		for _, oncall := range oncalls.OnCall {

			startDate, startErr := convertTimeBasedOnTimeZone(oncall.ShiftStart, timePattern, shift.TimeZone)
			if startErr != nil {
				logrus.Error("function sendNotification: Error in converting start time. username: ", receivers[i], "shift: ", shift.Name, "err: ", startErr)
				return fmt.Errorf("function sendNotification: Error in converting start time. username: %s. shift: %s, err: %w", receivers[i], shift.Name, startErr)
			}
			logrus.Debug("function sendNotification: preparing schedule list for user: ", oncall, ", start date is: ", startDate)

			endDate, endErr := convertTimeBasedOnTimeZone(oncall.ShiftEnd, timePattern, shift.TimeZone)
			if endErr != nil {
				logrus.Error("function sendNotification: Error in converting end time. username: ", receivers[i], "shift: ", shift.Name, "err: ", startErr)
				return fmt.Errorf("function sendNotification: Error in converting end time. username: %s. shift: %s, err: %w", receivers[i], shift.Name, startErr)
			}
			logrus.Debug("function sendNotification: preparing schedule list for user: ", oncall, ", end date is: ", endDate)

			oncallsList += startDate +
				" - " + endDate +
				": " + oncall.Username

			if oncall.Username == receivers[i] {
				oncallsList += " (You)"
			}
			oncallsList += "\n"

			logrus.Debug("function sendNotification: final shedule list is: ", oncallsList)

		}
		message := "Dear " + strings.Split(receivers[i], ".")[0] +
			", \nYour next on-call shift in schedule: " + shift.Name + " is about to begin shortly. Here's a summary of that shift:\n" +
			oncallsList + "\n" + "Best Regards,\nOncall-Notifier"

		logrus.Debug("function sendNotification: Message is ready for user: ", receivers[i], " ", message)

		logrus.Debug("function sendNotification: going to send email for ", receivers[i])
		reminderEmail := models.Email{
			Sender:   emailCredentials.Username,
			Receiver: receivers[i],
			Subject:  "Reminder: Your upcoming on-call shift for " + shift.Name,
			Body:     message,
		}
		if mailErr := email.SendMail(reminderEmail, *emailCredentials); mailErr != nil {
			logrus.Error("function sendNotification: Send Email failed. err: ", mailErr)
			return fmt.Errorf("function sendNotification: Send Email failed. err: %w", mailErr)
		}
		logrus.Debug("function sendNotification: email sent successfully.")
	}

	return nil
}

func uniqueNonEmptyElementsOf(s []string) []string {
	unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}

	return us
}

func convertTimeBasedOnTimeZone(currentTime string, timePattern string, targetTimezone string) (string, error) {
	dte, err := time.Parse(timePattern, currentTime)
	if err != nil {
		logrus.Error("function convertTimeBasedOnTimeZone: Error parsing entered datetime. err: ", err)
		return "", fmt.Errorf("function convertTimeBasedOnTimeZone: Error parsing entered datetime. err: %w", err)
	}
	logrus.Debug("function convertTimeBasedOnTimeZone: time parsed successfully.")

	localLoc, err := time.LoadLocation(targetTimezone)
	if err != nil {
		log.Fatal(`Failed to load location "Local"`)
		logrus.Error("function convertTimeBasedOnTimeZone: Failed to load location. err: ", err)
		return "", fmt.Errorf("function convertTimeBasedOnTimeZone: Failed to load location. err: %w", err)
	}
	logrus.Debug("function convertTimeBasedOnTimeZone: target location loaded successfully.")

	localDateTime := dte.In(localLoc)
	logrus.Debug("function convertTimeBasedOnTimeZone: converted time is: ", localDateTime)

	return localDateTime.String(), nil
}
