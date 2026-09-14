package helpers

// ValidateLuhn проверяет, является ли номер действительным по алгоритму Луна.
func ValidateLuhn(number string) bool {
	if number == "" {
		return false
	}

	var sum int
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		ch := number[i]

		if ch < '0' || ch > '9' {
			return false
		}

		digit := int(number[i] - '0')

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}
