package debug

import "fmt"

func Dump(n int, packet []byte) {
	for i := 0; i < n; i += 16 {
		msg := ""
		hex := ""
		ascii := ""
		for j := i; j < i+16 && j < n; j++ {
			b := string(packet[j])
			if packet[j] < 32 || packet[j] > 126 {
				b = "."
			}
			hex = fmt.Sprintf("%s%02X", hex, packet[j])
			ascii += fmt.Sprint(b)
			if (j+1)%2 == 0 {
				hex += " "
			}
		}
		space := 40 - len(hex)
		msg = hex
		for j := 0; j < space; j++ {
			msg += " "
		}
		msg += " " + ascii
		fmt.Println(msg)
	}
}
