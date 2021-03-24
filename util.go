package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	gomail "gopkg.in/mail.v2"
)

type response struct {
	OK bool `json:"ok"`
}

type uploadedResponse struct {
	response
	Path string `json:"path"`
}

func newUploadedResponse(path string) uploadedResponse {
	return uploadedResponse{response: response{OK: true}, Path: path}
}

type errorResponse struct {
	response
	Message string `json:"error"`
}

func newErrorResponse(err error) errorResponse {
	return errorResponse{response: response{OK: false}, Message: err.Error()}
}

func writeError(w http.ResponseWriter, err error) (int, error) {
	body := newErrorResponse(err)
	b, e := json.Marshal(body)

	// if an error is occured on marshaling, write empty value as response.
	if e != nil {
	    sendEmail(os.Getenv("ALERT_SMTP_OBJECT") + " ERROR marshaling", err.Error())
		return w.Write([]byte{})
	}
	sendEmail(os.Getenv("ALERT_SMTP_OBJECT") + " ERROR", err.Error())
	return w.Write(b)
}

func writeSuccess(w http.ResponseWriter, path string) (int, error) {
	body := newUploadedResponse(path)
	b, e := json.Marshal(body)
	// if an error is occured on marshaling, write empty value as response.
	if e != nil {
	    if os.Getenv("ALERT_SMTP_SEND_SUCCESS") == "true" {
            sendEmail(os.Getenv("ALERT_SMTP_OBJECT") + " SUCCESS ERROR marshaling", path)
        }
		return w.Write([]byte{})
	}
	if os.Getenv("ALERT_SMTP_SEND_SUCCESS") == "true" {
	    sendEmail(os.Getenv("ALERT_SMTP_OBJECT") + " SUCCESS", path)
	}

	return w.Write(b)
}

func getSize(content io.Seeker) (int64, error) {
	size, err := content.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}
	_, err = content.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func sendEmail(object string, message string) {
    m := gomail.NewMessage()

    // Set E-Mail sender
    m.SetHeader("From", os.Getenv("ALERT_SMTP_EMAIL_FROM"))

    // Set E-Mail receivers
    m.SetHeader("To", os.Getenv("ALERT_SMTP_EMAIL_TO"))

    // Set E-Mail subject
    m.SetHeader("Subject", object)

    // Set E-Mail body. You can set plain text or html with text/html
    m.SetBody("text/plain", message)

    // Settings for SMTP server
    d := gomail.NewDialer(os.Getenv("ALERT_SMTP_HOST"), 587, os.Getenv("ALERT_SMTP_EMAIL_FROM"), os.Getenv("ALERT_SMTP_EMAIL_PASSWORD"))

    // This is only needed when SSL/TLS certificate is not valid on server.
    // In production this should be set to false.
    //d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

    // Now send E-Mail
    if err := d.DialAndSend(m); err != nil {
        logger.WithError(err).Error("Failed to send alert error email")
    }
}