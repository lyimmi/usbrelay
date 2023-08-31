package usbrelay

const (
	// CfgVendorID 5824 = voti.nl
	CfgVendorID uint16 = 0x16c0

	// CfgDeviceID obdev's shared PID for HIDs
	CfgDeviceID uint16 = 0x05DF

	// RelayVendorName is the usbDevice's vendor's name
	RelayVendorName string = "www.dcttech.com"

	// RelayNamePref can be relay1... relay8
	RelayNamePref string = "USBRelay"

	// RelayIDLen length of "unique serial number" in the devices
	RelayIDLen int = 5
)

const (
	USB_TYPE_CLASS   = 0x01 << 5
	USB_RECIP_DEVICE = 0x00
	USB_ENDPOINT_IN  = 0x80
	USB_ENDPOINT_OUT = 0x00

	USBRQ_HID_GET_REPORT = 0x01
	USBRQ_HID_SET_REPORT = 0x09

	USB_HID_REPORT_TYPE_FEATURE = 3
)
