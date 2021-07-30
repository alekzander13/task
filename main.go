package main

import "time"

func main() {
	var c Configuration
	c.Init()

	srv := Server{
		Addr:         c.Port,
		IdleTimeout:  time.Duration(c.DeadlineSeconds) * time.Second,
		MaxReadBytes: c.MaxReadBytes,
	}

	srv.ListenAndServe()
}
