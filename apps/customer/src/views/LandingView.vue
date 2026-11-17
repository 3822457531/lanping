<template>
  <div class="page">
    <div class="card">
      <div class="mark">兰</div>
      <h1>兰瓶美业</h1>
      <p>客户无需安装 App。<br>核销时请打开短信中的专属链接，出示二维码给店员扫描即可。</p>
    </div>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh; display:grid; place-items:center; padding: 24px;
  background: linear-gradient(160deg, #f7d0da 0%, #f6f3f1 48%);
}
.card {
  width: min(420px, 100%); background:#fff; border-radius:20px; padding:28px 22px;
  text-align:center; border:1px solid #efe4e0;
}
.mark {
  width:48px; height:48px; margin:0 auto 14px; border-radius:14px;
  background:#d45a72; color:#fff; display:grid; place-items:center; font-weight:700; font-size:20px;
}
h1 { margin:0 0 10px; font-size:22px; }
p { margin:0; color:#8f857e; font-size:14px; line-height:1.7; }
</style>
