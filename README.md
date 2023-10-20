# StreamEth
Tool to ask for blocks from the Ethereum CL clients and evaluate block score.
It also has the capability to track some extra metrics from the Beacon Node, such as attestation reception.

# Execution
Please copy the env-example file into a .env file and fill the variables to adjust the execution to your desired scenario.

To run the project simply compile with `make build` and then execute with `make run`.
The variables in the .env file will be used.

Keep in mind the tool first needs to download the 64 previous blocks to the current head (for each beacon node), so as to build a history and score new blocks.

This tool has been tested on `go1.17.2 linux/amd64`

# Arguments

```

USAGE:
   live-metrics [command options] [arguments...]

OPTIONS:
   --bn-endpoints value  beacon node endpoints (label/endpoint,label/endpoint)
   --db-endpoint value   postgresql database endpoint: postgresql://user:password@localhost:5432/beaconchain 
   --db-workers value    10 (default: 1)
   --log-level value     info,debug,warn (default: info)
   --metrics value       proposals,attestations (default: proposals,attestations)
```

Please bear in mind the attestations metrics will increase the database size a lot


# Metrics
Several tables will be created when the tool is executed:
- t_att_metrics
- t_block_metrics
- t_missed_blocks
- t_score_metrics
- t_reorg_metrics

## Score Metrics

The main functionality of this tool is to ask the beacon node for beacon block proposals and then score them, taking into account the content. You may check the scoring algorithm [here](https://github.com/attestantio/vouch/blob/0c75ee8315dc4e5df85eb2aa09b4acc2b4436661/strategies/beaconblockproposal/best/score.go#L222)

The tool will ask, at every slot (at the head), one beacon block proposal to each of the configured beacon nodes. After this, the block will be analyzed and metrics will be stored in the table `t_score_metrics`

## Attestation Metrics

When activated through the metrics argument, the tool will subscribe to the attestation events of every beacon node. This is, to track every attestation seen by each of the beacon nodes, which would be stored in the table `t_att_metrics`

Keep in mind this metric can be very resource consuming (both CPU and Disk wise). One row per validator and epoch will be inserted in the table.

## Block Metrics

The tool will automatically subscribe to the head events of every beacon node and insert a new row in the table `t_block_metrics`, with the timestamp of the new block reception.
In case a head event is skipped, the tool will insert a new row in the table `t_missed_blocks`, as not receiving a head event in a slot is interpreted as a missed block.

## Reorg Metrics (experimental)

The tool can also subscribe to reorg events, if specified in the metrics argument. With this, the tool will insert a new row in the table `t_reorg_metrics` every time a reorg event is received from the beacon node.



