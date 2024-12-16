package job

import (
	"be/cache"
	"bytes"
	"context"
	"time"

	"github.com/nickalie/go-webpbin"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>TradingView Widget</title>
</head>
<body>
	<!-- TradingView Widget BEGIN -->
	<div class="tradingview-widget-container">
		<div class="tradingview-widget-container__widget"></div>
		<div class="tradingview-widget-copyright">
			<a href="https://www.tradingview.com/" rel="noopener nofollow" target="_blank">
				<span class="blue-text">Track all markets on TradingView</span>
			</a>
		</div>
		<script type="text/javascript" src="https://s3.tradingview.com/external-embedding/embed-widget-crypto-coins-heatmap.js" async>
		{
			"dataSource": "Crypto",
			"blockSize": "market_cap_calc",
			"blockColor": "change",
			"locale": "en",
			"symbolUrl": "",
			"colorTheme": "dark",
			"hasTopBar": false,
			"isDataSetEnabled": false,
			"isZoomEnabled": true,
			"hasSymbolTooltip": true,
			"isMonoSize": false,
			"width": "100%",
			"height": "100%"
		}
		</script>
	</div>
	<!-- TradingView Widget END -->
</body>
</html>
`

func SnapshotTradingViewHeatmap() ([]byte, error) {
	// Create a new context for Chromedp
	dockerURL := "wss://172.17.0.1:9222"
	allocatorContext, cancel := chromedp.NewRemoteAllocator(context.Background(), dockerURL)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()

	// Set up chromedp tasks
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://api.oneblock.vn/be/static/tradingview-heatmap-wg.html"),
		chromedp.WaitVisible(`div.tradingview-widget-container`, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Allow time for widget to load
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func StartJobSnapshotTradingViewHeatmap() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		tick := time.NewTicker(10 * time.Minute)
		fnJob := func() {
			image, err := SnapshotTradingViewHeatmap()
			if err != nil {
				zap.L().With(zap.Error(err)).Error("snapshot tradingview heatmap failed")
				return
			}
			webpOut := &bytes.Buffer{}
			mime := "image/png"
			err = webpbin.NewCWebP().Quality(100).Input(bytes.NewBuffer(image)).Output(webpOut).Run()
			if err != nil {
				zap.L().With(zap.Error(err)).Error("convert snapshot tradingview heatmap failed")
				webpOut = bytes.NewBuffer(image)
				mime = "image/webp"
			}
			imgData := &cache.ImageData{
				Data: webpOut.Bytes(),
				Mime: mime,
			}
			zap.L().Info("snapshot tradingview heatmap success")
			cache.CacheData.Store(cache.CacheDataTypeImage, imgData)
		}
		go fnJob()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				fnJob()

			}
		}
	}()
	return cancel

}
