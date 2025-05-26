![deployments_frequency](https://handler-badges.enpace.ch/df/Tiktai-badge)
[![lead_time_for_changes](https://handler-badges.enpace.ch/ltfc/Tiktai-badge)](https://handler-badges.enpace.ch/v1/Tiktai-badge/ltfc-stats)

<!---
TODO:
- Add description 
-->

Shows badges for DORA metrics from devlake

# Getting started

```bash
helm install dora-badge helm/dora-badge
```
Then you can add badges to your markdown 
```
[![lead_time_for_changes](https://<host>/v1/<devlake project>/ltfc)](https://<host>/v1/<devlake project>/ltfc-stats)

```