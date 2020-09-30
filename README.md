# goJSONparser
Json parses written in Go allowing not to use a pre defined struct (read values anonymously as byte slices)
default unmarshaling allows only for predefined data structure access making quick data exploration hard sometimes

## TODO
- [ ] take unicode characters larger than 8 bytes under account (current version assumes ASCII)
- [ ] change for loops into ones taking string sectioning into bytes not characters under account
- [ ] change json_object interface from slice into a map[string]anyjs
- [ ] add recursive parsing
