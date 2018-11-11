package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"os/exec"
	"path"
	"strings"

	"github.com/travis-ci/trvs-operator/pkg/apis/travisci/v1"
)

func NewTrvs(url string, key []byte, keychains Keychains) (*Trvs, error) {
	keys, err := ssh.NewPublicKeys("git", key, "")
	if err != nil {
		return nil, err
	}

	t := &Trvs{
		Path:          "/trvs",
		RepositoryURL: url,
		Keys:          keys,
		Keychains:     keychains,
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
	Keychains     Keychains
}

func (t *Trvs) initialize() error {
	if err := t.initializeRepo(); err != nil {
		return err
	}
	log.Info("initialized trvs repo")

	if err := t.installDeps(); err != nil {
		return err
	}
	log.Info("installed trvs dependencies")

	return nil
}

func (t *Trvs) initializeRepo() error {
	entry := log.WithFields(log.Fields{
		"path": t.Path,
		"url":  t.RepositoryURL,
	})

	var r *git.Repository
	r, err := git.PlainOpen(t.Path)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			// if the repository doesn't exist, make a fresh clone
			r, err = git.PlainClone(t.Path, false, &git.CloneOptions{
				URL:  t.RepositoryURL,
				Auth: t.Keys,
			})
			if err != nil {
				return err
			}

			entry.Info("cloned trvs")
		} else {
			return err
		}
	} else {
		// if the repository already existed, update it
		wt, err := r.Worktree()
		if err != nil {
			return err
		}

		if err = wt.Pull(&git.PullOptions{
			RemoteName: "origin",
			Auth:       t.Keys,
			Force:      true,
		}); err != nil {
			if err != git.NoErrAlreadyUpToDate {
				entry.WithError(err).Error("could not update trvs")
				return err
			}
		}

		entry.Info("updated trvs")
	}

	t.Repository = r
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
	var secrets map[string]interface{}
	var rawKeys bool

	if spec.File != "" {
		var k *Keychain
		if spec.IsPro {
			k = t.Keychains.Com
		} else {
			k = t.Keychains.Org
		}

		contents, err := k.ReadFile(spec.File)
		if err != nil {
			return nil, err
		}

		rawKeys = true
		secrets = make(map[string]interface{})
		secrets[spec.Key] = contents
	} else {
		var out bytes.Buffer
		cmd := exec.Command(t.exe(), "generate-config", "-n", "-f", "json", "-a", spec.App, "-e", spec.Environment)
		if spec.IsPro {
			cmd.Args = append(cmd.Args, "--pro")
		}
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return nil, err
		}

		if spec.Key != "" {
			rawKeys = true
			secrets = make(map[string]interface{})
			secrets[spec.Key] = out.Bytes()
		} else {
			if err := json.Unmarshal(out.Bytes(), &secrets); err != nil {
				return nil, err
			}
		}
	}

	return transformSecretData(spec, secrets, rawKeys), nil
}

func transformSecretData(spec v1.TrvsSecretSpec, data map[string]interface{}, rawKeys bool) map[string][]byte {
	newData := make(map[string][]byte)

	for k, v := range data {
		if !rawKeys {
			if spec.Prefix != "" {
				k = spec.Prefix + "_" + k
			}
			k = strings.ToUpper(k)
		}

		// K8s API handles base64 encoding the values, so just put the raw bytes in here
		if bytes, ok := v.([]byte); ok {
			newData[k] = bytes
		} else {
			newData[k] = []byte(fmt.Sprintf("%v", v))
		}
	}

	return newData
}
