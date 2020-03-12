package test_utils

func VerifyError(have error, want error) (bool, string, string) {
	switch {
	case have == nil && want == nil:
		return true, "", ""
	case have != nil && want == nil:
		return false, have.Error(), ""
	case have == nil && want != nil:
		return false, "", want.Error()
	default:
		return have.Error() == want.Error(), have.Error(), want.Error()
	}
}
