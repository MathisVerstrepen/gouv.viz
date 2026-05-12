#!/bin/sh
set -eu

root="web/assets/css"
src="$root/src"
out="$root/main.css"
tmp="$(mktemp)"

cleanup() {
	rm -f "$tmp"
}
trap cleanup EXIT

printf '/* Generated from web/assets/css/src by scripts/build-css.sh. Do not edit directly. */\n\n' > "$tmp"

for file in \
	"00-tokens.css" \
	"01-base.css" \
	"02-layout.css" \
	"03-utilities.css" \
	"components/shell-header.css" \
	"pages/shared-headings.css" \
	"pages/home.css" \
	"components/forms-filters.css" \
	"components/list-cards.css" \
	"components/chips-votes.css" \
	"components/pagination.css" \
	"pages/detail.css" \
	"visualizations/vote-breakdowns.css" \
	"visualizations/group-vote-chart.css" \
	"visualizations/group-deviation-chart.css" \
	"components/individual-votes.css" \
	"pages/deputy-detail.css" \
	"components/tables.css" \
	"components/footer.css" \
	"responsive.css" \
	"forced-colors.css"
do
	printf '/* %s */\n' "$file" >> "$tmp"
	cat "$src/$file" >> "$tmp"
	printf '\n' >> "$tmp"
done

if [ -f "$out" ] && cmp -s "$tmp" "$out"; then
	rm -f "$tmp"
else
	mv "$tmp" "$out"
fi

trap - EXIT
