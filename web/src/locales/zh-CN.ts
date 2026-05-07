export default {
  common: {
    save: "保存",
    cancel: "取消",
    confirm: "确认",
    delete: "删除",
    edit: "编辑",
    search: "搜索",
    phone: "手机号",
    username: "用户名",
    password: "密码",
  },
  login: {
    loginButton: "登录",
    forgotPassword: "忘记密码？",
    usernameValidator:
      "用户名必须以字母开头，只能包含字母和数字，长度为3到30个字符。",
    loginByPhone: "验证码登录",
    loginByPassword: "密码登录",
    phoneValidator: "请输入正确的手机号码",
    code: "验证码",
    codeValidator: "请输入正确的验证码",
    sendCode: "发送验证码",
    passwordValidator: "密码长度为6到30个字符",
  },
  home: {
    hero: {
      title: "Nexus",
      subtitle: "高性能内网穿透工具，轻松访问内网服务",
      learnMore: "了解更多",
      cta: "开始使用",
    },
    features: {
      title: "核心特性",
      subtitle: "强大的功能组合，满足您的内网穿透需求",
      nat: {
        title: "内网穿透",
        desc: "将内网服务安全地暴露到公网，无需公网 IP 或复杂配置",
      },
      secure: {
        title: "安全可靠",
        desc: "基于 HMAC 签名认证，多重鉴权机制保障通信安全",
      },
      manage: {
        title: "可视化管理",
        desc: "提供 Web 管理面板，轻松管理客户端、隧道和用户",
      },
      multi: {
        title: "多协议支持",
        desc: "支持 TCP 隧道和 V2Ray 代理，满足不同场景需求",
      },
      monitor: {
        title: "实时监控",
        desc: "实时查看系统状态、在线主机和资源使用情况",
      },
      cross: {
        title: "跨平台",
        desc: "客户端支持 Windows、Linux、macOS 等多种平台",
      },
    },
    howItWorks: {
      title: "工作原理",
      subtitle: "三步轻松实现内网穿透",
      step1: {
        title: "部署客户端",
        desc: "在内网机器上部署 Nexus 客户端，配置服务器地址和应用凭证",
      },
      step2: {
        title: "配置隧道",
        desc: "在管理面板中创建隧道，指定本地服务和公网监听端口",
      },
      step3: {
        title: "开始使用",
        desc: "通过公网地址访问内网服务，享受安全稳定的穿透体验",
      },
    },
    cta: {
      title: "现在就试试吧",
      subtitle: "一站式内网穿透解决方案，让您的服务触手可及",
      button: "前往控制台",
    },
  },
};
