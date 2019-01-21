package utils

import (
	"testing"
	"io/ioutil"
)

// TestIsFile test IsFile function.
func TestGetImageType(t *testing.T) {
	imgContent, err := ioutil.ReadFile("./data/timg.jpeg")
	if err != nil {
		t.Errorf("ReadFile err: %s", err.Error())
		return
	}

	imgType := GetImageType(imgContent)
	if imgType != "JPEG" {
		t.Errorf("GetImageType failed: Got [%s], expected [%s].", imgType, "JPEG")
		return
	}
}

// TestTranImage Test conversion image type.
func TestTranImage(t *testing.T) {
	srcFile := "./data/timg.jpeg"
	dstFile := "./data/timg.png"

	// parsing image
	img, err := DecodeImg(srcFile, "JPEG")
	if err != nil {
		t.Errorf("DecodeImg err: %s", err.Error())
		return
	}

	// create image
	err =  EncodeImage(dstFile, img, "PNG")
	if err != nil {
		t.Errorf("EncodeImage err: %s", err.Error())
		return
	}
}
