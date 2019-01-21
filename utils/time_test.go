package utils

import (
	"fmt"
	"testing"
)

// TestCurTime test CurTime function.
func TestCurTime(t *testing.T) {
	cur := CurTime()
	fmt.Println(cur)
}

// TestAddDate test AddDate function.
func TestAddDate(t *testing.T) {
	dt := AddDate(-1, -1, -1)
	dd := dt.Format("20060102")
	fmt.Println(dd)
}

// TestStrToTimeStamp test StrToTimeStamp function.
func TestStrToTimeStamp(t *testing.T) {
	strTime := "2017-07-06 16:27:28"
	st := StrToTimeStamp(strTime)
	if st != 1499329648 {
		t.Errorf("StrToTimeStamp err, Got %d, expected 1499329648", st)
		return
	}

	strTime = "20170706162728"
	fmtTime := "20060102150405"
	st = StrToTimeStamp(strTime, fmtTime)
	if st != 1499329648 {
		t.Errorf("StrToTimeStamp err, Got %d, expected 1499329648", st)
		return
	}
}

// TestTimeStampToStr test TimeStampToStr function.
func TestTimeStampToStr(t *testing.T) {
	var timeStamp int64 = 1499329648
	strTime := TimeStampToStr(timeStamp)
	if strTime != "2017-07-06 16:27:28" {
		t.Errorf("TimeStampToStr err, Got %s, expected 2017-07-06 16:27:28", strTime)
		return
	}

	fmtTime := "20060102150405"
	strTime = TimeStampToStr(timeStamp, fmtTime)
	if strTime != "20170706162728" {
		t.Errorf("TimeStampToStr err, Got %s, expected 20170706162728", strTime)
		return
	}
}

// TestTomrrowRest test TomrrowRest function.
func TestTomrrowRest(t *testing.T) {
	dd := TomrrowRest()
	fmt.Println(dd)
}
