package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func commitContainer(imageName string) {
	mntPath := "/home/joe/merged"
	imageTar := "/home/joe/" + imageName + ".tar"
	fmt.Println("Commit container imageTar:", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		log.Errorf("tar folder %s error %v", mntPath, err)
	}
}
