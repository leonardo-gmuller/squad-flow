<script setup lang="ts">
import Toast from './components/primevue/Toast'
import { RouterView } from 'vue-router'
</script>
<template>
  <div
    class="app-ambient"
    aria-hidden="true"
  ></div>
  <Toast
    class="workspace-toast"
    position="top-right"
    :auto-z-index="false"
  />
  <RouterView v-slot="{ Component, route }">
    <Transition
      name="session"
      mode="out-in"
    >
      <component
        :is="Component"
        :key="route.matched[0]?.path"
      />
    </Transition>
  </RouterView>
</template>
