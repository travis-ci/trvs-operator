package main

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"os/exec"
	"path"
	"strings"

	"github.com/travis-ci/trvs-operator/pkg/apis/travisci/v1"
)

func NewTrvs(url string, key []byte) (*Trvs, error) {
	keys, err := ssh.NewPublicKeys("git", key, "")
	if err != nil {
		return nil, err
	}

	t := &Trvs{
		Path:          "/trvs",
		RepositoryURL: url,
		Keys:          keys,
	}

	if err = t.initialize(); err != nil {
		return nil, err
	}

	return t, nil
}

type Trvs struct {
	Path          string
	RepositoryURL string
	Repository    *git.Repository
	Keys          *ssh.PublicKeys
}

func (t *Trvs) initialize() error {
	r, err := git.PlainClone(t.Path, false, &git.CloneOptions{
		URL:  t.RepositoryURL,
		Auth: t.Keys,
	})
	if err != nil {
		return err
	}

	t.Repository = r
	log.WithFields(log.Fields{
		"path": t.Path,
		"url":  t.RepositoryURL,
	}).Info("cloned trvs")

	if err = t.installDeps(); err != nil {
		return err
	}
	log.Info("installed trvs dependencies")

	return nil
}

func (t *Trvs) exe() string {
	return path.Join(t.Path, "bin", "trvs")
}

func (t *Trvs) installDeps() error {
	cmd := exec.Command("bundle", "install")
	cmd.Dir = t.Path
	return cmd.Run()
}

func (t *Trvs) Generate(spec v1.TrvsSecretSpec) (map[string][]byte, error) {
	var out bytes.Buffer
	cmd := exec.Command(t.exe(), "generate-config", "-n", "-f", "json", "-a", spec.App, "-e", spec.Environment)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var secrets map[string]string
	if err := json.Unmarshal(out.Bytes(), &secrets); err != nil {
		return nil, err
	}

	return transformSecretData(spec, secrets), nil
}

func transformSecretData(spec v1.TrvsSecretSpec, data map[string]string) map[string][]byte {
	newData := make(map[string][]byte)

	for k, v := range data {
		if spec.Prefix != "" {
			k = spec.Prefix + "_" + k
		}

		// K8s API handles base64 encoding it
		newData[strings.ToUpper(k)] = []byte(v)
	}

	return newData
}
