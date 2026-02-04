package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Nexa Protocol GUI")
	w.Resize(fyne.NewSize(600, 400))

	commandEntry := widget.NewEntry()
	commandEntry.SetPlaceHolder("مثال: PING أو FETCH homepage أو PUBLISH homepage مرحباً")

	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("النتيجة ستظهر هنا...")
	output.Wrapping = fyne.TextWrapWord

	dnsCheck := widget.NewCheck("استخدم DNS (.nexa)", nil)
	dnsCheck.SetChecked(false)

	sendBtn := widget.NewButton("إرسال", func() {
		cmd := commandEntry.Text
		if strings.TrimSpace(cmd) == "" {
			output.SetText("يرجى إدخال أمر!")
			return
		}
		var serverAddr string
		if dnsCheck.Checked {
			// استخراج الاسم
			parts := strings.Fields(cmd)
			if len(parts) < 2 {
				output.SetText("الأمر يحتاج اسم .nexa!")
				return
			}
			name := parts[1]
			addr, err := resolveDNS(name)
			if err != nil {
				output.SetText("فشل حل DNS: " + err.Error())
				return
			}
			serverAddr = addr
		} else {
			serverAddr = "localhost:1413"
		}
		resp, err := sendCommand(serverAddr, cmd)
		if err != nil {
			output.SetText("خطأ: " + err.Error())
			return
		}
		output.SetText(resp)
	})

	cont := container.NewVBox(
		widget.NewLabel("أدخل أمر Nexa أو استخدم DNS (.nexa):"),
		commandEntry,
		dnsCheck,
		sendBtn,
		output,
	)
	w.SetContent(cont)
	w.ShowAndRun()
}

func sendCommand(addr, cmd string) (string, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = fmt.Fprintf(conn, "%s\n", cmd)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	reader := make([]byte, 4096)
	for {
		n, err := conn.Read(reader)
		sb.Write(reader[:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return sb.String(), err
		}
		if strings.Contains(sb.String(), "---END---") {
			break
		}
	}
	return sb.String(), nil
}

func resolveDNS(name string) (string, error) {
	conn, err := tls.Dial("tcp", "localhost:1112", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "RESOLVE %s\n", name)
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	resp := string(buf[:n])
	parts := strings.SplitN(resp, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("استجابة DNS غير صالحة")
	}
	if parts[0] != "200" {
		return "", fmt.Errorf(resp)
	}
	body := parts[2]
	addr := strings.Split(body, "|")[0]
	return addr, nil
}
