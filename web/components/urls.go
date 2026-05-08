package components

import "net/url"

func ScrutinDetailURL(uid string) string {
	return "/scrutins/" + url.PathEscape(uid)
}
