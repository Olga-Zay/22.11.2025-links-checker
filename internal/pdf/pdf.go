package pdf

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"

	"links-checker/internal/domain"
)

func GenerateReport(tasks []*domain.LinkCheckTask) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	for _, task := range tasks {
		generateTaskLine(pdf, task)

		for _, link := range task.Links {
			generateLinkStatusLine(pdf, &link)
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func generateTaskLine(pdf *gofpdf.Fpdf, task *domain.LinkCheckTask) {
	pdf.Ln(4)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Task %d", task.ID))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
}

func generateLinkStatusLine(pdf *gofpdf.Fpdf, link *domain.Link) {
	line := fmt.Sprintf("%s - %s", link.URL, link.Status)
	pdf.Cell(40, 6, line)
	pdf.Ln(6)
}
