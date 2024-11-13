# Gotchas

## Restarting the cilium-agent results in a connectivity-loss
Error in the log after restart of cilium:
> Unable to restore endpoint, ignoring" endpointID=3411 error="Failed to re-allocate IP of endpoint: unable to reallocate 172.26.0.2 IPv4 address: provided IP is not in the valid range. The range of valid IPs is 10.1.0.0/16

