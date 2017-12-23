# flannel gw配置文件

cat > flannel.json << EOF
{
"Network": "10.254.0.0/16",
"SubnetLen": 26,
"SubnetMin": "10.254.0.64",
"SubnetMax": "10.254.250.192",
"Backend":
  {
    "Type": "host-gw"
  }
}
EOF