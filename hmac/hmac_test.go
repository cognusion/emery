package hmac

import (
	"crypto/rand"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_HMAC(t *testing.T) {

	Convey("Trivial random messages with trival random keys are signed and verified", t, func() {
		for i := 0; i < 100; i++ {
			key := randBytes(64)
			msg := randBytes(256)
			hash := signHMAC(msg, key)
			ok, err := verifyHMAC(msg, key, hash)
			So(ok, ShouldBeTrue)
			So(err, ShouldBeNil)
		}
	})

	Convey("Trivial random messages with trivial random keys are signed, but after trivial random changes to the messages the verifications fail", t, func() {
		for i := 0; i < 100; i++ {
			key := randBytes(64)
			msg := randBytes(256)
			hash := signHMAC(msg, key)
			ok, _ := verifyHMAC(msg[:randNumber(2)], key, hash)
			So(ok, ShouldBeFalse)
		}
	})
}

func randBytes(size int) []byte {
	bytes := make([]byte, size)

	rand.Read(bytes)
	return bytes
}

func randNumber(size int) int {
	charset := "0123456789"
	bytes := make([]byte, size)
	setLen := byte(len(charset))

	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = charset[v%setLen]
	}
	i, _ := strconv.Atoi(string(bytes))
	return i
}
