<template>
  <span
    class="brand app-logo"
    role="img"
    aria-label="SquadFlow"
  >
    <img
      class="brand-symbol"
      src="/logo.svg"
      alt=""
      width="64"
      height="64"
    />
    <span
      class="brand-word"
      aria-hidden="true"
      >squad<span>flow</span><b>.</b></span
    >
  </span>
</template>

<style>
.login-card .app-logo {
  margin-bottom: 24px;
}
.login-card .brand-word {
  display: block;
  color: var(--sf-text);
}
.login-card .brand-word > span,
.login-card .brand-word > b {
  color: var(--sf-accent);
}
</style>
