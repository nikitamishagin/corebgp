# CoreBGP

**THE PROJECT IS STILL UNDER DEVELOPMENT**

## Description

CoreBGP is a full-fledged service for managing BGP announcements within your infrastructure. It provides a REST API,
health checks, IPAM integration, and high availability storage. The service architecture is designed for modularity,
high availability, and seamless integration with other systems. CoreBGP enables you to add custom BGP announcements to
balance the load across your application replicas.

### Advantages

What sets CoreBGP apart from other solutions? Existing implementations like GoBGP and ExaBGP are excellent products.
However, they do not address all the requirements needed to act as a comprehensive BGP announcement controller in your
infrastructure. CoreBGP fills this gap by combining the best practices and providing a complete feature set for working
with BGP.

## Components

The project is composed of several modules (microservices):

- etcd (support for other databases is planned in the future);
- API server;
- Updater;
- GoBGP;
- Checker (multiple implementations are possible);
- IPAM plugin (optional component).

### ETCD

ETCD is used as a high availability storage cluster to save BGP announcements, their states, and the states of service
components. To ensure high availability, it is recommended to use a multi-node cluster. Deployments with a single ETCD
instance are only suitable for development and testing purposes. The choice of ETCD as the primary database is based on
the following reasons:

- It is a noSQL database. Since the service has a narrow focus, most of the database will consist of a single table—BGP
  announcements. Using relational databases in this case seems excessive. However, support for other databases through
  adapters is planned in the future.
- Simplifies the deployment of fault-tolerant clusters without requiring manual load balancing. In some database
  systems, clusterization can be a complex task that complicates the service deployment process.
- Supports writing to any node in the cluster. This feature is particularly valuable in high-load systems, as writing to
  a local database node (on the same host) can significantly reduce network communication delays.
- Easy to configure. ETCD can be launched with just a few startup arguments, delivering excellent performance out of the
  box without fine-tuning.
- Actively maintained and continuously evolving.

### API server

_readme in progress..._
