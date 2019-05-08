package utils

import (
	"image"
	"os"
	"image/jpeg"
	"fmt"
	"image/png"
	"image/gif"
	"strings"
)

// GetImageType returns the type of image.
// Currently only supports PNG, JPEG, GIF, BMP, and other types return empty strings.
func GetImageType(imgContent []byte) string {
	if len(imgContent) < 4 {
		return ""
	}
	
	if imgContent[0] == 137 && imgContent[1] == 80 {
		return "PNG"
	} else if imgContent[0] == 255 && imgContent[1] == 216 {
		return "JPEG"
	} else if imgContent[0] == 71 && imgContent[1] == 73 && imgContent[2] == 70 && (imgContent[4] == 55 || imgContent[4] == 57) {
		return "GIF"
	} else if imgContent[0] == 66 && imgContent[1] == 77 {
		return "BMP"
	}

	return ""
}

// DecodeImg parsing image.
// Currently only supports PNG, JPEG and GIF.
func DecodeImg(imgFile string, imgType string) (image.Image, error) {
	// open image
	f, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	// decode
	var img image.Image
	iType := strings.ToUpper(imgType)
	if iType == "JPEG" {
		img, err = jpeg.Decode(f)
	} else if iType == "PNG" {
		img, err = png.Decode(f)
	} else if iType == "GIF" {
		img, err = gif.Decode(f)
	} else {
		return nil, fmt.Errorf("Image type [%s] is not supported", imgType)
	}
	if err != nil {
		return nil, err
	}

	return img, nil
}

// EncodeImage create image.
// Currently only supports PNG, JPEG and GIF.
// option: JPEG-[1,100], GIF-[1,256]
func EncodeImage(imgFile string, img image.Image, imgType string, option ...int) error {
	// mkdir
	err := MkDir(imgFile, os.ModeDir|0755, true)
	if err != nil {
		return err
	}

	// open image
	f, err := os.OpenFile(imgFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if nil != err {
		return err
	}
	defer f.Close()

	// write image
	imgType = strings.ToUpper(imgType)
	if imgType == "PNG" {
		return png.Encode(f, img)
	} else if imgType == "JPEG" {
		option = append(option, 90)
		op := option[0]
		return jpeg.Encode(f, img, &jpeg.Options{op})
	} else if imgType == "GIF" {
		option = append(option, 100)
		op := option[0]
		return gif.Encode(f, img, &gif.Options{NumColors: op})
	}

	return fmt.Errorf("Image type [%s] is not supported", imgType)
}
