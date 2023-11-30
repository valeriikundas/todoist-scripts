package toggl

type TelegramTimeoutError struct {
}

func (e *TelegramTimeoutError) Error() string {
	return "timed out waiting for telegram reply"
}

func (e *TelegramTimeoutError) Is(target error) bool {
	_, ok := target.(*TelegramTimeoutError)
	return ok
}
