package actionlint

// RuleBranding is a rule to check branding information of actions
type RuleBranding struct {
	RuleBase
}

// NewRuleBranding creates new RuleBranding instance
func NewRuleBranding() *RuleBranding {
	return &RuleBranding{
		RuleBase: RuleBase{
			name: "branding",
			desc: "Checks for valid branding information of actions",
		},
	}
}

// VisitActionPre is callback when visiting Action node before visiting its children.
func (rule *RuleBranding) VisitActionPre(n *Action) error {
	if n.Branding == nil { // Branding is optional
		return nil
	}

	// TODO auto-extract it from GitHub docs
	valid_colors := []string{"white", "yellow", "blue", "green", "orange", "red", "purple", "gray-dark"}
	if n.Branding.Color != nil && !contains(valid_colors, n.Branding.Color.Value) {
		rule.Errorf(n.Branding.Color.Pos, "invalid color %q. valid colors are %v", n.Branding.Color.Value, valid_colors)
	}

	// TODO auto-extract it from GitHub docs
	valid_icons := []string{"activity", "airplay", "alert-circle", "alert-octagon", "alert-triangle", "align-center", "align-justify", "align-left", "align-right", "anchor", "aperture", "archive", "arrow-down-circle", "arrow-down-left", "arrow-down-right", "arrow-down", "arrow-left-circle", "arrow-left", "arrow-right-circle", "arrow-right", "arrow-up-circle", "arrow-up-left", "arrow-up-right", "arrow-up", "at-sign", "award", "bar-chart-2", "bar-chart", "battery-charging", "battery", "bell-off", "bell", "bluetooth", "bold", "book-open", "book", "bookmark", "box", "briefcase", "calendar", "camera-off", "camera", "cast", "check-circle", "check-square", "check", "chevron-down", "chevron-left", "chevron-right", "chevron-up", "chevrons-down", "chevrons-left", "chevrons-right", "chevrons-up", "circle", "clipboard", "clock", "cloud-drizzle", "cloud-lightning", "cloud-off", "cloud-rain", "cloud-snow", "cloud", "code", "command", "compass", "copy", "corner-down-left", "corner-down-right", "corner-left-down", "corner-left-up", "corner-right-down", "corner-right-up", "corner-up-left", "corner-up-right", "cpu", "credit-card", "crop", "crosshair", "database", "delete", "disc", "dollar-sign", "download-cloud", "download", "droplet", "edit-2", "edit-3", "edit", "external-link", "eye-off", "eye", "fast-forward", "feather", "file-minus", "file-plus", "file-text", "file", "film", "filter", "flag", "folder-minus", "folder-plus", "folder", "gift", "git-branch", "git-commit", "git-merge", "git-pull-request", "globe", "grid", "hard-drive", "hash", "headphones", "heart", "help-circle", "home", "image", "inbox", "info", "italic", "layers", "layout", "life-buoy", "link-2", "link", "list", "loader", "lock", "log-in", "log-out", "mail", "map-pin", "map", "maximize-2", "maximize", "menu", "message-circle", "message-square", "mic-off", "mic", "minimize-2", "minimize", "minus-circle", "minus-square", "minus", "monitor", "moon", "more-horizontal", "more-vertical", "move", "music", "navigation-2", "navigation", "octagon", "package", "paperclip", "pause-circle", "pause", "percent", "phone-call", "phone-forwarded", "phone-incoming", "phone-missed", "phone-off", "phone-outgoing", "phone", "pie-chart", "play-circle", "play", "plus-circle", "plus-square", "plus", "pocket", "power", "printer", "radio", "refresh-ccw", "refresh-cw", "repeat", "rewind", "rotate-ccw", "rotate-cw", "rss", "save", "scissors", "search", "send", "server", "settings", "share-2", "share", "shield-off", "shield", "shopping-bag", "shopping-cart", "shuffle", "sidebar", "skip-back", "skip-forward", "slash", "sliders", "smartphone", "speaker", "square", "star", "stop-circle", "sun", "sunrise", "sunset", "tablet", "tag", "target", "terminal", "thermometer", "thumbs-down", "thumbs-up", "toggle-left", "toggle-right", "trash-2", "trash", "trending-down", "trending-up", "triangle", "truck", "tv", "type", "umbrella", "underline", "unlock", "upload-cloud", "upload", "user-check", "user-minus", "user-plus", "user-x", "user", "users", "video-off", "video", "voicemail", "volume-1", "volume-2", "volume-x", "volume", "watch", "wifi-off", "wifi", "wind", "x-circle", "x-square", "x", "zap-off", "zap", "zoom-in", "zoom-out"}
	if n.Branding.Icon != nil && !contains(valid_icons, n.Branding.Icon.Value) {
		rule.Errorf(n.Branding.Icon.Pos, "invalid icon %q", n.Branding.Icon.Value)
	}
	return nil
}
