package tat

type HookMessageJSON struct {
	Action   string  `json:"action"`
	Username string  `json:"username"`
	Message  Message `json:"message"`
}

type Hook struct {
	message HookMessageJSON
}
