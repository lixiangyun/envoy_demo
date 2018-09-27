## 基于 centos7_1804 版本

## 安装依赖库

```
yum makecache
yum install -y gcc gcc-c++  kernel-devel kernel-headers kernel.x86_64 net-tools
yum install -y numactl-devel.x86_64 numactl-libs.x86_64
yum install -y libpcap.x86_64 libpcap-devel.x86_64
yum install -y pciutils
```

## clone mtcp

```
git clone https://github.com/eunyoung14/mtcp.git
```

## mtcp

```
bash setup_mtcp_env.sh
```

