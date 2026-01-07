package version

import (
	"fmt"
	"strconv"
	"time"
)

// Version is the application version. Can be overridden at build time via:
//
//	go build -ldflags "-X winsbygroup.com/regserver/internal/version.Version=1.2.3"
var Version = "1.0"

// RepoURL is the project repository URL. Can be overridden at build time via:
//
//	go build -ldflags "-X winsbygroup.com/regserver/internal/version.RepoURL=https://github.com/yourfork/regserver"
var RepoURL = "https://github.com/winsbygroup/regserver"

// Banner prints identifying information about the server.
func Banner() string {
	y := strconv.Itoa(time.Now().Year())
	copyright := "Copyright 2025-" + y + " Winsby Group LLC. All rights reserved."

	return fmt.Sprintf("%s\nRegserver (v%s)\n%s\n", product(), Version, copyright)
}

func product() string {
	// http://patorjk.com/software/taag/#p=display&f=Standard&t=Regserver
	// it includes back ticks, which makes this more difficult (replace with `+"`"+`).

	const s = `
  ____                                         
 |  _ \ ___  __ _ ___  ___ _ ____   _____ _ __ 
 | |_) / _ \/ _` + "`" + ` / __|/ _ \ '__\ \ / / _ \ '__|
 |  _ <  __/ (_| \__ \  __/ |   \ V /  __/ |   
 |_| \_\___|\__, |___/\___|_|    \_/ \___|_|   
            |___/
`
	return s
}
