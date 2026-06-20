# Rules

- 每次更新完后台或前端相关服务后，必须对相应的 Systemd 服务（如 s-ui, sing-box, s-ui-agent）进行彻底重启（例如 `systemctl restart <service_name>`），以确保最新代码和配置完全加载生效。
