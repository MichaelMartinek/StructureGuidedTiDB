# Structure Guided TiDB

This repository is a fork of TiDB extended with an implementation of structure guided query optimization. The new optimization option for acyclic and nearly acyclic conjunctive queries uses query decompositions and Yannakakis' algorithm in order to define a good join order and eliminate tuples not contributing to the end result of the query, before the actual join.

## Dependencies

TiDB is a go project, therefore requiring a valid go installation. We used go version 1.20.

Query decompositions are created through the go library [BalancedGo](https://github.com/cem-okulmus/BalancedGo).


## Running the DBMS

To run or debug the DBMS, follow the instructions in the [original README file](README_ORIG.md).

OR

See the [Get Started](https://pingcap.github.io/tidb-dev-guide/get-started/introduction.html) chapter of [TiDB Development Guide](https://pingcap.github.io/tidb-dev-guide/index.html).

### Patching a Cluster with TiUP

- Build the DBMS: `make`
- Package the binary (from the bin folder): `tar -cvzf out.tar.gz  ./tidb-server`
- Patch all TiDB instances in the database cluster: `tiup cluster patch <cluster_name> <path_to_packagedBinary>  -R tidb`


## Evaluations


Information regarding conducted evaluations are contained the following repository: [StructureGuidedTiDBEvaluation](https://github.com/MichaelMartinek/StructureGuidedTiDBEvaluation) 
