# UVY Design adoption

This application vendors [UVY Design v0.1.0](https://github.com/UnitVectorY-Labs/uvy-design/tree/12f8e3d964785d78628346516351fe4dac583cef), source commit `12f8e3d964785d78628346516351fe4dac583cef`. The [style guide](https://github.com/UnitVectorY-Labs/uvy-design/blob/12f8e3d964785d78628346516351fe4dac583cef/docs/style-guide.md) defines the visual contract; application CSS remains the adapter.

Current [UVY Design guidance](https://github.com/UnitVectorY-Labs/uvy-design/tree/7091d87) supersedes the older style guide for visual decisions: cards and panels use neutral borders without colored top or left accents. The files below remain pinned from the earlier release.

## Pinned assets

| Verbatim shared file | SHA-256 |
| --- | --- |
| `assets/uvy-tokens.css` | `0e97baaece886e0015c4238eaf7dbed9ea68f330c0f50729311c43b4a5a98533` |
| `assets/uvy-elements.css` | `b2f92292dd9317bb770be1a4c745c9a00b81b6cef39cbcac54d2fe192135b092` |
| `assets/uvy-theme.js` | `b105c0ddec8c19b6a380a33ffa5f9f558ca49baa477921a1dee9596d77bc61ed` |

`assets/vendor/htmx-4.0.0.min.js` is the existing 4.0.0 dependency served locally instead of through unpkg; its SHA-384 remains `BvJpBiO8Kh31EqtJe5DRIeWrHWnCGkwytKs9NKFi86Hhw96dEqdEMzZDeK9iEGTc`. No major-version migration or extra frontend framework is introduced. Runtime assets and full license notices are copied to generated output by the Go generator.

Official Tabler outline paths for menu-2, moon, sun, search and brand-github are inlined in `templates/index.html` and `templates/nav.html`. They originate in the release's icon snapshot, Tabler commit `ed177a085b652fe04ec4f37ced867c94621e7840`; they are decorative SVG with accessible action names on their controls. Provider artwork remains the app's existing local Auth0/Okta/Salesforce/Ping marks. UVY Design MIT, Tabler MIT and HTMX Zero-Clause BSD notices are under `assets/licenses/` and are copied into output.

## Adapter and preserved behavior

`assets/style.css` maps existing header, sidebar, content, endpoints, provider tables, aliases and active/inactive key history to the shared tokens. Standard density uses 16px panels and 12px/16px table padding; sidebar and provider tables use compact whitespace without reducing essential technical values below 14px. Mobile key rows preserve KID, use, algorithm/curve, key length and observation fields in compact groups; no columns or record sections are removed. Active/history semantics remain distinct in both themes.

The sidebar/content layout, provider filters, home explanation, endpoint order, aliases, provider-evidence matrix and key history remain in place. Full pages and fragments continue to share the service template. Initial navigation is now generated from the same nav partial into each full page as well as `nav.html`, improving no-JavaScript entry without changing destinations. The existing mobile toggle remains, with expanded state, nested-link close and Escape handling. HTMX 4 title/focus/current-link sync follows its colon-separated events.

Appearance defaults to the OS. A top-right native light/dark button stores a host-only `uvy-theme` cookie using the unmodified shared controller. No localStorage or cross-domain appearance state is introduced. Footer links identify this project's source/license and the observer data source. The footer is in the existing flex document shell, never a fixed overlay. System fonts and inline/local assets require no external runtime fetches.

## Verification and screenshot evidence

Run `go test ./...`, `go vet ./...`, `go run main.go -validate`, and generate with `WEBSITE` plus optional `JWKS_OBSERVER_PATH`. Browser verification covers full/fragment navigation, direct entry, history/title/focus, filters, theme persistence/system default, no-JS browse, provider artwork, 320/390px fit and no external runtime requests. Screenshot evidence is attached to the PR from an immutable design-repository evidence commit, not committed to this application.

The screenshot baseline is origin/main `0d3f9cd57149ab38b515c472409d2a281b190e57`; both stages use the same complete catalog and observer checkout `20ff65dc8b1a6b25ae0fbc50ed7979c9556cc222`. Before has no dark mode; both requested OS themes are captured faithfully rather than fabricating one. Independent scroll panes require supplementary captures for provider-evidence and inactive-history content below the first viewport. Screenshots do not establish a full accessibility certification.

## Updating

Select a reviewed future commit, verify file checksums, replace the shared files verbatim, retain notices, and update this record. Review adapter diffs and repeat application/browser/screenshot checks. Do not retrieve mutable upstream CSS/JS at runtime or overwrite application rules with the design prototype's layout stylesheet.
