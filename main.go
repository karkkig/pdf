package main

import (
	"fmt"
	"log"
	"os"
	"time"

	cdpdf "pdf_poc/chromedp"
	mapdf "pdf_poc/maroto"
	pcpdf "pdf_poc/pdfcpu"
)

func main() {

	os.MkdirAll("output", os.ModePerm)

	fmt.Println("Generating chromedp PDF...")

	invoiceHTML, err := cdpdf.ReadHTML("chromedp/template/invoice.html")
	if err != nil {
		panic(err)
	}
	err = cdpdf.GeneratePDF(invoiceHTML, "output/invoice-chromedp.pdf")
	if err != nil {
		panic(err)
	}

	badgeHTML, err := cdpdf.ReadHTML("chromedp/template/badge.html")
	if err != nil {
		panic(err)
	}
	err = cdpdf.GeneratePDF(badgeHTML, "output/badge-chromedp.pdf")
	if err != nil {
		panic(err)
	}

	agreementHTML, err := cdpdf.ReadHTML("chromedp/template/agreement.html")
	if err != nil {
		panic(err)
	}
	err = cdpdf.GeneratePDF(agreementHTML, "output/agreement-chromedp.pdf")
	if err != nil {
		panic(err)
	}

	fmt.Println("Generating maroto PDF...")

	// ── 1. Invoice ────────────────────────────────────────────────────────────
	invoiceData := mapdf.InvoiceData{
		Number: "123456",
		Date:   time.Date(2030, 5, 24, 0, 0, 0, 0, time.UTC),
		Provider: mapdf.PartyInfo{
			Name:    "STUDIO SHODWE",
			Address: "123 Anywhere St., Any City",
			City:    "ST 12345",
			Phone:   "+123-456-7890",
			Email:   "hello@reallygreatsite.com",
		},
		Client: mapdf.PartyInfo{
			Name:    "Rachel Beaudry",
			Address: "123 Anywhere St., Any City",
			City:    "ST 12345",
			Phone:   "+123-456-7890",
			Email:   "hello@reallygreatsite.com",
		},
		Items: []mapdf.InvoiceItem{
			{"Service 1", 100.00, 1},
			{"Service 2", 150.00, 1},
			{"Service 3", 200.00, 1},
		},
		TaxRate: 0.06,
		Notes:   "Payment is due within 15 days\nof receiving this invoice.",
		PaymentMethod: mapdf.PaymentInfo{
			Bank:          "Borcelle Bank",
			AccountName:   "Studio Shodwe",
			AccountNumber: "1234567890",
		},
		PreparedBy: mapdf.PreparedByInfo{
			Name:  "Benjamin Shah",
			Title: "Sales Administrator, Studio Shodwe",
		},
	}
	if err := mapdf.GenerateInvoice("output/invoice-maroto.pdf", invoiceData); err != nil {
		log.Fatal("invoice:", err)
	}

	// ── 2 & 3. Services Agreement (p.1) + Payment Plan (p.2) — single PDF ────
	fullAgreement := mapdf.FullAgreementData{
		// Page 1 — Services Agreement
		State:           "California",
		Day:             "1st",
		Month:           "January",
		Year:            "25",
		ProviderName:    "Acme Services Inc.",
		ProviderAddress: "123 Main Street, Los Angeles, CA 90001",
		BuyerName:       "Globex Corp.",
		BuyerAddress:    "456 Market Street, San Francisco, CA 94105",
		Services: []mapdf.ServiceItem{
			{"Web Development", "2", "5,000"},
			{"UI/UX Design", "1", "3,000"},
			{"SEO Package", "3", "1,200"},
			{"Maintenance (monthly)", "6", "800"},
		},
		PurchasePrice: "20,000",
		Notes:         "All work will be delivered digitally. Revisions are limited to 3 rounds per project.",

		// Page 2 — Payment Plan
		Payer:           "John Doe",
		Payee:           "Jane Smith",
		Product:         "Web Development Services",
		AmountPerPeriod: "$500",
		Interval:        "month",
		TotalAmount:     "$3,000",
		Payments: []mapdf.PaymentEntry{
			{"1 Feb 2025", "$500"},
			{"1 Mar 2025", "$500"},
			{"1 Apr 2025", "$500"},
			{"1 May 2025", "$500"},
			{"1 Jun 2025", "$500"},
			{"1 Jul 2025", "$500"},
		},
		LateFee:         "$50",
		BounceFee:       "$75",
		LenderAction:    "contact a debt collection service",
		TermsConditions: "No refunds after work commencement. Disputes subject to California jurisdiction.",
	}
	if err := mapdf.GenerateFullAgreement("output/agreement-maroto.pdf", fullAgreement); err != nil {
		log.Fatal("full agreement:", err)
	}

	badgeData := mapdf.GPayBadgeData{
		BusinessName: "Your Business Name",
		PhoneNumber:  "+91 12345 67890",
		UPIHandle:    "12345 67890@yhh",
		//QRCodePath:   "qr.png", // Set to your QR image path, e.g. "qr.png"
		QRContent: "upi://pay?pa=1234567890@yhh&pn=Your+Business+Name&cu=INR",
	}
	if err := mapdf.GenerateGPayBadge("output/badge-maroto.pdf", badgeData); err != nil {
		log.Fatal("gpay badge:", err)
	}

	fmt.Println("Generating pdfcpu PDF...")

	if err := pcpdf.GenerateInvoice("output/invoice-pdfcpu.pdf"); err != nil {
		log.Fatalf("Failed to generate invoice: %v", err)
	}

	// Generate Payment Badge PDF
	if err := pcpdf.GeneratePaymentBadge("output/badge-pdfcpu.pdf"); err != nil {
		log.Fatalf("Failed to generate payment badge: %v", err)
	}

	// Generate Two-Page Services Agreement PDF
	if err := pcpdf.GenerateAgreement("output/agreement-pdfcpu.pdf"); err != nil {
		log.Fatalf("Failed to generate agreement: %v", err)
	}

	fmt.Println("✅ All PDFs generated successfully!")
}
