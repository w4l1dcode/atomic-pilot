# atomic-pilot

This Go program automates the execution of Atomic Red Team tests, which are small, portable tests aligned with the MITRE ATT&CK framework. Its main purpose is to systematically execute each Technique, Tactic, and Procedure (TTP) associated with different threat actors. During each run, the program selects a new threat actor along with their corresponding TTPs. By running the program on a device with your security tools, you can test your detection capabilities. Results are logged via Slack. The program is compatible with both macOS and Windows.

# Prerequisites

### PowerShell (for running Atomic Red Team tests)
For Macos:
```shell
brew install powershell/tap/powershell
```

## Running
First create a yaml file, such as `config.yml`:
```yaml
log:
  level: INFO

slack:
  url: ""
```

And now run the program from source code:
```shell
% make
go run ./cmd/... -config=dev.yml
INFO[0000] set log level                                 fields.level=debug
INFO[0005] Techniques used by OilRig (AKA OilRig, COBALT GYPSY, IRN2, APT34, Helix Kitten, Evasive Serpens, Hazel Sandstorm, EUROPIUM): 
INFO[0005] T1555.004                                    
INFO[0005] T1082                                        
INFO[0005] T1003.001                                    
INFO[0005] T1008   
```

## Building

```shell
% make build
```

