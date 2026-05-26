// CreateQRWithLogo generates a QR code using the WithLogo option.
// The QR is generated at native resolution based on dimension, logo is
// pre-resized to exactly 1/5 of the QR canvas, then saved as dimensionxdimension.
func CreateQRWithLogo(content string, logoURL string, dimension int) error {
	fmt.Printf("Creating QR code with content: %s, logoURL: %s, dimension: %d\n",
		content, logoURL, dimension)

	qr, err := qrcode.New(content)
	if err != nil {
		fmt.Printf("create qrcode failed: %v\n", err)
		return err
	}

	// Derive module width from dimension and actual QR version
	// so the generated image is already at native resolution
	version := qrVersionFromContent(content)
	modules := (version-1)*4 + 21
	qrWidth := uint8(dimension / modules)
	if qrWidth < 1 {
		qrWidth = 1
	}

	fmt.Printf("QR version: %d, modules: %d, qrWidth per module: %d\n", version, modules, qrWidth)

	var options []standard.ImageOption

	if logoURL != "" {
		if err = UrlGet(logoURL); err != nil {
			fmt.Printf("failed to fetch logo: %v\n", err)
			return err
		}

		// Pre-resize logo to exactly 1/5 of the native QR canvas size
		nativeQRSize := int(qrWidth) * modules
		targetLogoSize := nativeQRSize / 5

		fmt.Printf("Native QR size: %dpx, target logo size: %dpx\n", nativeQRSize, targetLogoSize)

		if err = resizeLogoToTarget("logo1.jpg", targetLogoSize); err != nil {
			fmt.Printf("failed to resize logo: %v\n", err)
			return err
		}

		options = []standard.ImageOption{
			standard.WithLogoImageFileJPEG("logo1.jpg"),
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(0),
		}
	} else {
		options = []standard.ImageOption{
			standard.WithQRWidth(qrWidth),
			standard.WithBorderWidth(0),
		}
	}

	writer, err := standard.New("qrcode_with_logo.png", options...)
	if err != nil {
		fmt.Printf("create writer failed: %v\n", err)
		return err
	}
	defer writer.Close()

	if err = qr.Save(writer); err != nil {
		fmt.Printf("save qrcode failed: %v\n", err)
		return err
	}

	// Final resize to exact dimension in case of rounding differences
	if err = resizeImage("qrcode_with_logo.png", dimension); err != nil {
		fmt.Printf("resize failed: %v\n", err)
		return err
	}

	return nil
}
