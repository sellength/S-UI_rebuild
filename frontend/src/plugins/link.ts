import { Hysteria, Hysteria2, InTypes, Inbound, Naive, Shadowsocks, TUIC, Trojan, VLESS, VMess, AnyTLS } from "@/types/inbounds"
import { HTTP, WebSocket, gRPC, HTTPUpgrade, Transport, TrspTypes } from "@/types/transport"
import RandomUtil from "./randomUtil"
import Data from "@/store/modules/data"

export interface Link {
  type: "local" | "external" | "sub"
  remark?: string
  uri: string
}

function utf8ToBase64(utf8String: string): string {
  const encodedUtf8 = encodeURIComponent(utf8String).replace(/%([0-9A-F]{2})/g, (_, p1) => String.fromCharCode(parseInt(p1, 16)));
  return btoa(encodedUtf8);
}

export namespace LinkUtil {
  export function linkGenerator(user: string | any, inbound: Inbound, tlsClient: any = {}, addrs: any[] = []): string[] {
    const name = typeof user === 'string' ? user : (user as any)?.name || ''
    switch(inbound.type){
      case InTypes.Shadowsocks:
        return shadowsocksLink(name,<Shadowsocks>inbound, addrs)
      case InTypes.Naive:
        return naiveLink(name,<Naive>inbound, addrs, tlsClient)
      case InTypes.Hysteria:
        return hysteriaLink(name,<Hysteria>inbound, addrs, tlsClient)
      case InTypes.Hysteria2:
        return hysteria2Link(name,<Hysteria2>inbound, addrs, tlsClient)
      case InTypes.TUIC:
        return tuicLink(name,<TUIC>inbound, addrs, tlsClient)
      case InTypes.VLESS:
        return vlessLink(name,<VLESS>inbound, addrs, tlsClient)
      case InTypes.Trojan:
        return trojanLink(name,<Trojan>inbound, addrs, tlsClient)
      case InTypes.VMess:
        return vmessLink(name,<VMess>inbound, addrs, tlsClient)
      case InTypes.AnyTLS:
        return anytlsLink(name,<AnyTLS>inbound, addrs, tlsClient)
    }
    return []
  }

  function shadowsocksLink(user: string, inbound: Shadowsocks, addrs: any[]): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let userPass = u?.password
    if (!userPass) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      userPass = client?.config?.shadowsocks?.password
    }
    if (!userPass) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.password) { userPass = cfg.password; break }
          if (cfg && cfg.uuid) { userPass = cfg.uuid; break }
          if (cfg && cfg.auth_str) { userPass = cfg.auth_str; break }
        }
      }
    }
    if (!userPass) userPass = RandomUtil.randomSeq(10)

    const password = [userPass]
    if (inbound.method.startsWith('2022')) password.push(inbound.password)
    const params = {
      tfo: inbound.tcp_fast_open? 1 : null,
      network: inbound.network?? null
    }

    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`ss://${utf8ToBase64(inbound.method + ':' + password.join(':'))}@${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries(params)){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`ss://${utf8ToBase64(inbound.method + ':' + password.join(':'))}@${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries(params)){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }

  function hysteriaLink(user: string, inbound: Hysteria, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let auth = u?.auth_str
    if (!auth) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      auth = client?.config?.hysteria?.auth_str
    }
    if (!auth) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.auth_str) { auth = cfg.auth_str; break }
          if (cfg && cfg.password) { auth = cfg.password; break }
          if (cfg && cfg.uuid) { auth = cfg.uuid; break }
        }
      }
    }
    if (!auth) auth = RandomUtil.randomSeq(10)

    const params = {
      upmbps: inbound.up_mbps?? null,
      downmbps: inbound.down_mbps?? null,
      auth: auth?? null,
      peer: inbound.tls.server_name?? null,
      alpn: inbound.tls.alpn?.join(',')?? null,
      obfsParam: inbound.obfs?? null,
      fastopen: inbound.tcp_fast_open? 1 : 0,
      insecure: tlsClient?.insecure ? 1 : null
    }

    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`hysteria://${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries(params)){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`hysteria://${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries(params)){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        if (a.server_name?.length>0) {
          uri.searchParams.set('peer', a.server_name)
        } else {
          inbound.tls.server_name ? uri.searchParams.set('peer', inbound.tls.server_name) : uri.searchParams.delete('peer')
        }
        if (a.insecure) {
          uri.searchParams.set('insecure', '1')
        } else {
          tlsClient.insecure ? uri.searchParams.set('insecure', '1') : uri.searchParams.delete('insecure')
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }

  function hysteria2Link(user: string, inbound: Hysteria2, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let password = u?.password
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      password = client?.config?.hysteria2?.password
    }
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.password) { password = cfg.password; break }
          if (cfg && cfg.uuid) { password = cfg.uuid; break }
          if (cfg && cfg.auth_str) { password = cfg.auth_str; break }
        }
      }
    }
    if (!password) password = RandomUtil.randomSeq(10)

    const params = {
      upmbps: inbound.up_mbps?? null,
      downmbps: inbound.down_mbps?? null,
      sni: inbound.tls.server_name?? null,
      alpn: inbound.tls.alpn?.join(',')?? null,
      obfs: inbound.obfs?.type?? null,
      'obfs-password': inbound.obfs?.password?? null,
      fastopen: inbound.tcp_fast_open? 1 : 0,
      insecure: tlsClient?.insecure ? 1 : null
    }

    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`hysteria2://${password}@${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries(params)){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`hysteria2://${password}@${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries(params)){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        if (a.server_name?.length>0) {
          uri.searchParams.set('sni', a.server_name)
        } else {
          inbound.tls.server_name ? uri.searchParams.set('sni', inbound.tls.server_name) : uri.searchParams.delete('sni')
        }
        if (a.insecure) {
          uri.searchParams.set('insecure', '1')
        } else {
          tlsClient.insecure ? uri.searchParams.set('insecure', '1') : uri.searchParams.delete('insecure')
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }

  function naiveLink(user: string, inbound: Naive, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.username == user)
    let password = u?.password
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      password = client?.config?.naive?.password
    }
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.password) { password = cfg.password; break }
          if (cfg && cfg.uuid) { password = cfg.uuid; break }
          if (cfg && cfg.auth_str) { password = cfg.auth_str; break }
        }
      }
    }
    if (!password) password = RandomUtil.randomSeq(10)

    let links = <string[]>[]
    if (addrs.length == 0) {
      const params = {
        padding: 1,
        peer: inbound.tls.server_name?? null,
        alpn: inbound.tls.alpn?.join(',')?? null,
        tfo: inbound.tcp_fast_open? 1 : 0,
        allowInsecure: tlsClient?.insecure ? 1 : null
      }
      const uri = `http2://${utf8ToBase64(user + ":" + password + "@" + location.hostname + ":" + inbound.listen_port)}`
      const paramsArray = []
      for (const [key, value] of Object.entries(params)){
        if (value) {
          paramsArray.push(`${key}=${encodeURIComponent(value.toString())}`)
        }
      }
      links.push(uri.toString() + "?" + paramsArray.join('&') + "#" + inbound.tag)
    } else {
      addrs.forEach(a => {
        const params = {
          padding: 1,
          peer: a.server_name?.length>0 ? a.server_name : inbound.tls.server_name?? null,
          alpn: inbound.tls.alpn?.join(',')?? null,
          tfo: inbound.tcp_fast_open? 1 : 0,
          allowInsecure: a.insecure ? 1 : tlsClient?.insecure ? 1 : null
        }
        const uri = `http2://${utf8ToBase64(user + ":" + password + "@" + a.server + ":" + a.server_port)}`
        const paramsArray = []
        for (const [key, value] of Object.entries(params)){
          if (value) {
            paramsArray.push(`${key}=${encodeURIComponent(value.toString())}`)
          }
        }
        links.push(uri.toString() + "?" + paramsArray.join('&') + "#" + encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag))
      })
    }
    return links
  }

  function tuicLink(user: string, inbound: TUIC, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let uuid = u?.uuid
    let password = u?.password
    if (!uuid || !password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (!uuid) uuid = client?.config?.tuic?.uuid
      if (!password) password = client?.config?.tuic?.password
    }
    if (!uuid || !password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.uuid && !uuid) uuid = cfg.uuid
          if (cfg && cfg.password && !password) password = cfg.password
          if (cfg && cfg.auth_str && !password) password = cfg.auth_str
        }
      }
    }
    if (!uuid) uuid = RandomUtil.randomUUID()
    if (!password) password = RandomUtil.randomSeq(10)

    const params = {
      sni: inbound.tls.server_name?? null,
      alpn: inbound.tls.alpn?.join(',')?? null,
      congestion_control: inbound.congestion_control?? null,
      allowInsecure: tlsClient?.insecure ? 1 : null,
      disable_sni: tlsClient?.disable_sni ? 1 : null
    }

    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`tuic://${uuid}:${password}@${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries(params)){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`tuic://${uuid}:${password}@${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries(params)){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        if (a.server_name?.length>0) {
          uri.searchParams.set('sni', a.server_name)
        } else {
          inbound.tls.server_name ? uri.searchParams.set('sni', inbound.tls.server_name) : uri.searchParams.delete('sni')
        }
        if (a.insecure) {
          uri.searchParams.set('allowInsecure', '1')
        } else {
          tlsClient.insecure ? uri.searchParams.set('allowInsecure', '1') : uri.searchParams.delete('allowInsecure')
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }

  function getTransportParams(t:Transport): any {
    if (Object.keys(t).length == 0) return {}

    const params = {
      host: <string|null>'',
      path: <string|null>'',
      serviceName: <string|null>'',
    }
    switch (t.type){
      case TrspTypes.HTTP:
        const th = <HTTP>t
        params.host = th.host?.join(',')?? null
        params.path = th.path?? null
        break
      case TrspTypes.WebSocket:
        const tw = <WebSocket>t
        params.path = tw.path?? null
        params.host = tw.headers?.Host?? null
        break
      case TrspTypes.gRPC:
        const tg = <gRPC>t
        params.serviceName = tg.service_name?? null
        break
      case TrspTypes.HTTPUpgrade:
        const tu = <HTTPUpgrade>t
        params.host = tu.host?? null
        params.path = tu.path?? null
        break
    }

    return params
  }

  function vlessLink(user: string, inbound: VLESS, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let uuid = u?.uuid
    let flow = u?.flow
    if (!uuid) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      uuid = client?.config?.vless?.uuid
      flow = client?.config?.vless?.flow
    }
    if (!uuid) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.uuid) { uuid = cfg.uuid; break }
          if (cfg && cfg.password) { uuid = cfg.password; break }
          if (cfg && cfg.auth_str) { uuid = cfg.auth_str; break }
        }
      }
    }
    if (!uuid) uuid = RandomUtil.randomUUID()

    const transport = <Transport>inbound.transport
    const tParams = getTransportParams(transport)

    const params = {
      type: transport?.type?? 'tcp',
      security: inbound.tls?.enabled? inbound.tls?.reality?.enabled ? 'reality' : 'tls' : null,
      alpn: inbound.tls?.alpn?.join(',')?? null,
      sni: inbound.tls?.server_name?? null,
      flow: inbound.tls?.enabled ? flow?? null : null,
      allowInsecure: tlsClient?.insecure ? 1 : null,
      fp: tlsClient?.utls?.enabled ? tlsClient.utls.fingerprint : null,
      pbk: tlsClient?.reality?.public_key?? null,
      sid: inbound.tls?.reality?.enabled ? (inbound.tls?.reality?.short_id?.length>0 ?  inbound.tls.reality.short_id[RandomUtil.randomInt(inbound.tls.reality.short_id.length)] : null) : null
    }
    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`vless://${uuid}@${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries({...params, ...tParams})){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`vless://${uuid}@${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries({...params, ...tParams})){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        if (a.tls != undefined){
          if (a.tls) {
            uri.searchParams.set('security','tls')
          } else {
            uri.searchParams.delete('security')
            uri.searchParams.delete('sni')
            uri.searchParams.delete('alpn')
            uri.searchParams.delete('allowInsecure')
          }
        }
        if (a.server_name?.length>0) {
          uri.searchParams.set('sni', a.server_name)
        } else {
          inbound.tls?.server_name ? uri.searchParams.set('sni', inbound.tls.server_name) : uri.searchParams.delete('sni')
        }
        if (a.insecure) {
          uri.searchParams.set('allowInsecure', '1')
        } else {
          tlsClient.insecure ? uri.searchParams.set('allowInsecure', '1') : uri.searchParams.delete('allowInsecure')
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }

  function trojanLink(user: string, inbound: Trojan, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let password = u?.password
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      password = client?.config?.trojan?.password
    }
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.password) { password = cfg.password; break }
          if (cfg && cfg.uuid) { password = cfg.uuid; break }
          if (cfg && cfg.auth_str) { password = cfg.auth_str; break }
        }
      }
    }
    if (!password) password = RandomUtil.randomSeq(10)

    const transport = <Transport>inbound.transport
    const tParams = getTransportParams(transport)

    const params = {
      type: transport?.type?? 'tcp',
      security: inbound.tls?.enabled? inbound.tls?.reality?.enabled ? 'reality' : 'tls' : null,
      alpn: inbound.tls?.alpn?.join(',')?? null,
      sni: inbound.tls?.server_name?? null,
      allowInsecure: tlsClient?.insecure ? 1 : null,
      fp: tlsClient?.utls?.enabled ? tlsClient.utls.fingerprint : null,
      pbk: tlsClient?.reality?.public_key?? null,
      sid: inbound.tls?.reality?.enabled ? (inbound.tls?.reality?.short_id?.length>0 ?  inbound.tls.reality.short_id[RandomUtil.randomInt(inbound.tls.reality.short_id.length)] : null) : null
    }

    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`trojan://${password}@${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries({...params, ...tParams})){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`trojan://${password}@${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries({...params, ...tParams})){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        if (a.tls != undefined){
          if (a.tls) {
            uri.searchParams.set('security','tls')
          } else {
            uri.searchParams.delete('security')
            uri.searchParams.delete('sni')
            uri.searchParams.delete('alpn')
            uri.searchParams.delete('allowInsecure')
          }
        }
        if (a.server_name?.length>0) {
          uri.searchParams.set('sni', a.server_name)
        } else {
          inbound.tls?.server_name ? uri.searchParams.set('sni', inbound.tls.server_name) : uri.searchParams.delete('sni')
        }
        if (a.insecure) {
          uri.searchParams.set('allowInsecure', '1')
        } else {
          tlsClient.insecure ? uri.searchParams.set('allowInsecure', '1') : uri.searchParams.delete('allowInsecure')
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }

  function vmessLink(user: string, inbound: VMess, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let uuid = u?.uuid
    let alterId = u?.alterId ?? 0
    if (!uuid) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      uuid = client?.config?.vmess?.uuid
      alterId = client?.config?.vmess?.alterId ?? 0
    }
    if (!uuid) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.uuid) { uuid = cfg.uuid; break }
          if (cfg && cfg.password) { uuid = cfg.password; break }
          if (cfg && cfg.auth_str) { uuid = cfg.auth_str; break }
        }
      }
    }
    if (!uuid) uuid = RandomUtil.randomUUID()

    const transport = <Transport>inbound.transport
    const tParams = getTransportParams(transport)
    if (transport.type == TrspTypes.gRPC) tParams.path = tParams.serviceName

    const params = {
      v: 2,
      add: location.hostname,
      aid: alterId,
      host:	tParams.host?? undefined,
      id: uuid,
      net: transport?.type == undefined || transport?.type == 'http' ? 'tcp' : transport.type,
      type: transport?.type == 'http' ? 'http' : undefined,
      path:	tParams.path?? undefined,
      port:	inbound.listen_port,
      ps:	inbound.tag,
      sni: inbound.tls.server_name?? undefined,
      tls: Object.keys(inbound.tls).length>0? 'tls' : 'none',
      allowInsecure: tlsClient?.insecure ? 1 : undefined
    }
    let links = <string[]>[]
    if (addrs.length == 0) {
      links.push('vmess://' + utf8ToBase64(JSON.stringify(params, null, 2)))
    } else {
      addrs.forEach(a => {
        let newParams = {...params}
        newParams.add = a.server
        newParams.port = a.server_port
        if (a.tls != undefined){
          if (a.tls) {
            newParams.tls = 'tls'
          } else {
            newParams.tls = 'none'
            delete newParams.sni
            delete newParams.allowInsecure
          }
        }
        if (a.server_name?.length>0) {
          newParams.sni = a.server_name
        }
        if (a.insecure) {
          newParams.allowInsecure = 1
        }
        newParams.ps = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push('vmess://' + utf8ToBase64(JSON.stringify(newParams, null, 2)))
      })
    }
    return links
  }

  function anytlsLink(user: string, inbound: AnyTLS, addrs: any[], tlsClient: any): string[] {
    const u = inbound.users?.find(i => i && i.name == user)
    let password = u?.password
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      password = client?.config?.anytls?.password
    }
    if (!password) {
      const client = Data().clients?.find((c: any) => c.name === user) as any
      if (client?.config) {
        for (const key in client.config) {
          const cfg = client.config[key]
          if (cfg && cfg.password) { password = cfg.password; break }
          if (cfg && cfg.uuid) { password = cfg.uuid; break }
          if (cfg && cfg.auth_str) { password = cfg.auth_str; break }
        }
      }
    }
    if (!password) password = RandomUtil.randomSeq(10)

    const params = {
      sni: inbound.tls.server_name?? null,
      alpn: inbound.tls.alpn?.join(',')?? null,
      insecure: tlsClient?.insecure ? 1 : null
    }

    let links = <string[]>[]
    if (addrs.length == 0) {
      const uri = new URL(`anytls://${password}@${location.hostname}:${inbound.listen_port}`)
      for (const [key, value] of Object.entries(params)){
        if (value) {
          uri.searchParams.set(key, value.toString())
        }
      }
      uri.hash = encodeURIComponent(inbound.tag)
      links.push(uri.toString())
    } else {
      addrs.forEach(a => {
        const uri = new URL(`anytls://${password}@${a.server}:${a.server_port}`)
        for (const [key, value] of Object.entries(params)){
          if (value) {
            uri.searchParams.set(key, value.toString())
          }
        }
        if (a.server_name?.length>0) {
          uri.searchParams.set('sni', a.server_name)
        } else {
          inbound.tls.server_name ? uri.searchParams.set('sni', inbound.tls.server_name) : uri.searchParams.delete('sni')
        }
        if (a.insecure) {
          uri.searchParams.set('insecure', '1')
        } else {
          tlsClient.insecure ? uri.searchParams.set('insecure', '1') : uri.searchParams.delete('insecure')
        }
        uri.hash = encodeURIComponent(a.remark ? inbound.tag + a.remark : inbound.tag)
        links.push(uri.toString())
      })
    }
    return links
  }
}