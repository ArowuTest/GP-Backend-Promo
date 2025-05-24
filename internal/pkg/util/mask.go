package util

// MaskMSISDN masks an MSISDN (show only first 3 and last 3 digits)
func MaskMSISDN(msisdn string) string {
	if len(msisdn) <= 6 {
		return msisdn
	}
	
	prefix := msisdn[:3]
	suffix := msisdn[len(msisdn)-3:]
	masked := prefix + "****" + suffix
	
	return masked
}
