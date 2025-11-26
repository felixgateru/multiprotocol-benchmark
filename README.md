# Multiprotocol benchmarking tool

## Supported protocols

1. MQTT
2. COAP
3. HTTP
4. Websockests

## How to run

1. The program requires that you have a registered account on Magistrala Cloud that is verified. Please provide the details in the relevant env variable `MG_USERNAME` and `MG_PASSWORD`

2. If you have a domain you would like to use please provide it in the env varible,`DOMAIN_ID` if the env variable is left empty a domain will be created.

3. If you have channels and clients you would like to use please fill in the provison.toml file provided and set the `CONFIG_TOML` env variable to the file location. If this env variable is left blank then clients and channels will be created and numbers are determined by the `CHANNEL_COUNT` and `CLIENT_COUNT` variables

4. You can enable all or one of the protocols to benchmark. A `MESSAGE_COUNT` and `MESSAGE_DELAY` vaiable are provided for each protocol. You can set the `MESSAGE_DELAY` variable to zero to get the best possible benchmarks for the protocol
