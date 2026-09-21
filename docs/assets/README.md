# Repository preview assets

`gwc-homepage.png` is a browser screenshot of the public homepage at
<https://girlswhocodehunter.org/>, captured on September 11, 2026 (UTC). The URL is
configured in `hosting/aws/main.go`. The desktop viewport was 1440 × 1000 at
2× scale, producing a 2880 × 2000 PNG after page images and fonts loaded.

To refresh the preview, capture the same URL and viewport in a clean browser
session, check that the homepage has loaded completely, and replace the PNG.
Keep the root README's relative image path unchanged.

The backend diagram is reused from
[`backend-architecture.svg`](../../infrastructure/legacy/docs/architecture/assets/backend-architecture.svg),
with its editable
[Excalidraw source](../../infrastructure/legacy/docs/architecture/assets/backend-architecture.excalidraw)
preserved alongside it. The detailed hosting diagram remains in `hosting/docs/`.
