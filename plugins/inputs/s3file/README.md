# S3 File Input Plugin

The file plugin parses the **complete** contents of an S3 file **every interval** using
the selected [input data format][].

You may want to use this with the -once switch, so that Telegraf only runs once.

**Note:** If you wish to parse only newly appended lines use the [tail][] input
plugin instead.

### Configuration:

```toml
[[inputs.s3file]]
  ## Files to parse each interval.  Accept standard unix glob matching rules,
  ## as well as ** to match recursive files and directories.
  bucket = "bucket"
  keys = ["/tmp/metrics.out"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Name a tag containing the name of the file the data was parsed from.  Leave empty
  ## to disable.
  # include_key_in_tag = false
```

[input data format]: /docs/DATA_FORMATS_INPUT.md
[tail]: /plugins/inputs/tail
