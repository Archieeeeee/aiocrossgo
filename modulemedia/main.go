package main

import "AioCrossGo/media"

func main() {
	media.Init()

	media.ReloadCfg()
	media.StartUI()
}
