<template>
  <span class="pill">
    <slot />
  </span>
</template>

<style scoped>
.pill {
  font-size: 11px;
  font-weight: 600;
  color: var(--user-accent);
  background: var(--user-icon-bg);
  border-radius: 20px;
  padding: 6px 10px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

@media (max-width: 420px) {
  .pill {
    padding: 5px 7px;
  }
}
</style>
