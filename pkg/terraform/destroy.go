package terraform

import (
	"encoding/json"
)

func Destroy(workDirectory string, config *Config) error {
	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return err
	}

	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = tec.AddFile("config.tf.json", configBytes)
		if err != nil {
			return err
		}
	}

	result, _, err := tec.Init()
	if err != nil {
		return err
	}

	err = result.Wait()
	if err != nil {
		return err
	}

	result, _, err = tec.Destroy(nil)
	if err != nil {
		return err
	}

	return result.Wait()
}
