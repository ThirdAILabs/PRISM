package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"prism/prism/api"
	"prism/prism/services"
	"prism/prism/services/auth"
	"time"
)

var minUpdateInterval = 15 * 24 * 60 * 60 // 15 days in seconds.

type AuthorReportUpdateNotifierData struct {
	EmailID string `json:"email_id"`
}

type AuthorReportUpdateNotifier struct {
	BaseURL  string
	notifier *services.EmailMessenger
}

func NewAuthorReportUpdateNotifier(BaseUrl string, notifier *services.EmailMessenger) *AuthorReportUpdateNotifier {
	return &AuthorReportUpdateNotifier{
		BaseURL:  BaseUrl,
		notifier: notifier,
	}
}

func (h *AuthorReportUpdateNotifier) Validate(data []byte, interval int) error {
	if interval < minUpdateInterval {
		return fmt.Errorf("interval must be greater than or equal to %d days", minUpdateInterval/(24*60*60))
	}

	return nil
}

func (h *AuthorReportUpdateNotifier) Run(report api.Report, data []byte, lastRanAt time.Time) error {
	var hookData AuthorReportUpdateNotifierData
	if err := json.Unmarshal(data, &hookData); err != nil {
		return fmt.Errorf("failed to unmarshal hook data: %w", err)
	}

	newFlags := make([]api.Flag, 0)
	for _, flags := range report.Content {
		for _, flag := range flags {
			date, dateValid := flag.Date()
			if dateValid && date.After(lastRanAt) {
				newFlags = append(newFlags, flag)
			}
		}
	}

	if len(newFlags) == 0 {
		// No new flags since the last run
		return nil
	}
	authorReportEndpoint, err := url.JoinPath(h.BaseURL, "report", report.Id.String())
	if err != nil {
		return fmt.Errorf("failed to join URL path: %w", err)
	}
	return h.notify(hookData.EmailID, report.AuthorName, authorReportEndpoint, newFlags)
}

func (h *AuthorReportUpdateNotifier) Type() string {
	return "AuthorReportUpdateNotifier"
}

func (h *AuthorReportUpdateNotifier) CreateHookData(r *http.Request, payload []byte, interval int) (hookData []byte, err error) {
	email_id, err := auth.GetUserEmail(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get email id: %w", err)
	}
	obj := AuthorReportUpdateNotifierData{
		EmailID: email_id,
	}

	hookData, err = json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal hook data: %w", err)
	}
	return hookData, nil
}

func (h *AuthorReportUpdateNotifier) renderReportUpdateTemplate(authorName string, flagCount map[string]int, authorReportEndpoint string) (string, error) {
	const tmpl = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Author report update</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			line-height: 1.6;
			color: #333333;
			margin: 0;
			padding: 0;
			background-color: #f4f4f4;
		}
		.email-container {
			max-width: 600px;
			margin: 0 auto;
			background-color: #ffffff;
		}
		.header {
			padding: 20px;
			text-align: center;
		}
		.header img {
			max-width: 180px;
		}
		.content {
			padding: 20px;
			background-color: #ffffff;
		}
		h1 {
			color: #64b6f7;
			margin-top: 0;
		}
		.button {
			display: inline-block;
			background-color: #2196f3;
			color: white !important;
			padding: 12px 20px;
			text-decoration: none;
			border-radius: 4px;
			margin: 20px 0;
		}
		.footer {
			background-color: #f4f4f4;
			padding: 15px;
			text-align: center;
			font-size: 12px;
			color: #666666;
		}
		.footer a {
			color: #2196f3;
		}
		/* Email-compatible list styling */
		.flags-list {
			width: 100%;
			border-collapse: collapse;
			margin: 20px 0;
		}
		.flag-row {
			border-bottom: 1px dotted #e0e0e0;
		}
		.flag-title {
			padding: 8px 0;
			text-align: left;
		}
		.flag-count {
			font-weight: bold;
			color: #2196f3;
			padding: 8px 0;
			text-align: right;
			white-space: nowrap;
		}
		.table-header {
			background-color: #f8f9fa;
			font-weight: bold;
			color: #333333;
			border-bottom: 2px solid #e0e0e0;
		}
		.table-header td {
			padding: 10px 0;
		}
	</style>
</head>
<body>
	<div class="email-container">
		<div class="header">
			<img src="https://i.ibb.co/T97f2vP/prism-logo.png" alt="Prism Logo">
		</div>
		<div class="content">
			<h1>Author Report Update</h1>
			<p>We've identified new potential flags in our recent monitoring of author <b>{{.AuthorName}}</b>.</p>
			
			<table class="flags-list">
				<tr class="table-header">
					<td class="flag-title">Flag Category</td>
					<td class="flag-count">Newly Founded</td>
				</tr>
				{{range .Flags}}
					<tr class="flag-row">
						<td class="flag-title">{{.Title}}</td>
						<td class="flag-count">{{.Count}}</td>
					</tr>
				{{end}}
			</table>
			
			<a href="{{.AuthorReportEndpoint}}" class="button">View Full Report</a>
			
			<p>Thank you for choosing Prism. We're excited to be part of your journey</p>
			<p>Best regards,<br>ThirdAI Team</p>
		</div>
		<div class="footer">
			<p>If you have any questions, please email us at <a href="mailto:support@thirdai.com">support@thirdai.com</a></p>
		</div>
	</div>
</body>
</html>`

	type tmp struct {
		Title string
		Count int
	}

	flags := make([]tmp, 0)
	for title, count := range flagCount {
		if count > 0 {
			flags = append(flags, tmp{Title: title, Count: count})
		}
	}

	data := struct {
		AuthorName           string
		Flags                []tmp
		AuthorReportEndpoint string
	}{
		AuthorName:           authorName,
		Flags:                flags,
		AuthorReportEndpoint: authorReportEndpoint,
	}

	t := template.Must(template.New("authorUpdateEmail").Parse(tmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

func (h *AuthorReportUpdateNotifier) notify(receipientEmail, authorName, authorReportEndpoint string, flags []api.Flag) error {
	flagInfo := map[string]string{
		api.TalentContractType:               "Talent Contracts",
		api.AssociationsWithDeniedEntityType: "Funding from Denied Entities",
		api.HighRiskFunderType:               "High Risk Funding Sources",
		api.AuthorAffiliationType:            "Affiliations with High Risk Foreign Institutes",
		api.PotentialAuthorAffiliationType:   "Appointments at High Risk Foreign Institutes",
		api.MiscHighRiskAssociationType:      "Miscellaneous High Risk Connections",
		api.CoauthorAffiliationType:          "Co-authors are affiliated with Entities of Concern",
	}

	flagCount := make(map[string]int)
	for _, flag := range flags {
		flagCount[flagInfo[flag.Type()]]++
	}

	emailContent, err := h.renderReportUpdateTemplate(authorName, flagCount, authorReportEndpoint)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	err = h.notifier.SendEmail("prism@thirdai.com", receipientEmail, "Author Report Update", "", emailContent)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
