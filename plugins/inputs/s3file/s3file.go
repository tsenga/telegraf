package file

import (
	"fmt"
	"io/ioutil"
	"errors"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/dimchansky/utfbom"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type S3File struct {
	
	CharacterEncoding string   `toml:"character_encoding"`
	
	Keys              []string `toml:"keys"`
	Region            string   `toml:"region"`
	Bucket            string   `toml:"bucket"`
	KeyTag            string   `toml:"key_tag"`

	parser            parsers.Parser

	decoder   *encoding.Decoder

	client	          *s3.S3
}

const sampleConfig = `
  ## Files to parse each interval.  Accept standard unix glob matching rules,
  ## as well as ** to match recursive files and directories.
  bucket = "bucket"
  keys = ["/prefix/filename"]
  region = "eu-west-2"

  ## Name a tag containing the name of the file the data was parsed from.  Leave empty
  ## to disable.
  # include_key_in_tag = false

  ## Character encoding to use when interpreting the file contents.  Invalid
  ## characters are replaced using the unicode replacement character.  When set
  ## to the empty string the data is not decoded to text.
  ##   ex: character_encoding = "utf-8"
  ##       character_encoding = "utf-16le"
  ##       character_encoding = "utf-16be"
  ##       character_encoding = ""
  # character_encoding = ""

  ## The dataformat to be read from files
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

// SampleConfig returns the default configuration of the Input
func (f *S3File) SampleConfig() string {
	return sampleConfig
}

func (f *S3File) Description() string {
	return "Parse a complete file each interval"
}

func (f *S3File) Init() error {
	var err error
	f.decoder, err = encoding.NewDecoder(f.CharacterEncoding)
	return err
}

func (f *S3File) Gather(acc telegraf.Accumulator) error {
	initS3Client(f)

	for _, k := range f.Keys {
		metrics, err := f.readMetric(k)
		if err != nil {
			return err
		}

		for _, m := range metrics {
			if f.KeyTag != "" {
				m.AddTag(f.KeyTag, filepath.Base(k))
			}
			acc.AddMetric(m)
		}
	}
	return nil
}

func initS3Client(s3file *S3File) error {
	if (s3file.client == nil) {
		if (s3file.Region == "") {
			return errors.New("Expecting a S3 region to connect to")
		}

		sess, err := session.NewSession(&aws.Config {
			Region: aws.String(s3file.Region) },
		)

		if err != nil {
			return err
		}

		s3file.client = s3.New(sess)
	}

	return nil
}

func (f *S3File) SetParser(p parsers.Parser) {
	f.parser = p
}

func (f *S3File) readMetric(key string) ([]telegraf.Metric, error) {

	resp, err := f.client.GetObject(&s3.GetObjectInput {
		Bucket: aws.String(f.Bucket),
		Key: aws.String(f.Keys[0])	})


	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	r, _ := utfbom.Skip(f.decoder.Reader(resp.Body))
	fileContents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("E! Error file: %v could not be read, %s", key, err)
	}
	return f.parser.Parse(fileContents)
}

func init() {
	inputs.Add("s3file", func() telegraf.Input {
		return &S3File{}
	})
}
