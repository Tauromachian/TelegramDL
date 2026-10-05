<script setup>
import { ref } from 'vue'
import { ChevronRight } from '../icons'

defineProps({
  title: { type: String, required: true },
  description: { type: String, default: '' }
})

const isOpen = ref(false)
</script>

<template>
  <section class="accordion-block" :class="{ open: isOpen }">
    <button
      type="button"
      class="accordion-header"
      :aria-expanded="isOpen"
      @click="isOpen = !isOpen"
    >
      <span class="accordion-icon">
        <slot name="icon"></slot>
      </span>

      <span class="accordion-title">
        <p>
          <strong>{{ title }}</strong>
        </p>
        <p>
          <small>
            {{ description }}
          </small>
        </p>
      </span>

      <ChevronRight class="accordion-chevron" :size="16" />
    </button>

    <div v-show="isOpen" class="accordion-body">
      <slot></slot>
    </div>
  </section>
</template>

<style scoped>
.accordion-title > * {
  margin: 0;
}

.accordion-title strong {
  font:
    600 13px 'Space Grotesk',
    sans-serif;
}
.accordion-title small {
  font-size: 11px;
  color: var(--user-text-dim);
}

.accordion-block {
  border: 1px solid var(--user-border);
  border-radius: 14px;
  background: var(--user-bg-base);
  overflow: hidden;
  transition: border-color 0.2s;
}

.accordion-block.open {
  border-color: var(--user-border-light);
}

.accordion-header {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 14px 16px;
  background: transparent;
  border: 0;
  cursor: pointer;
  text-align: left;
  color: #dbe7f5;
  transition: background 0.2s;
}
.accordion-header:hover {
  background: var(--user-surface-light);
}
.accordion-header:focus-visible {
  outline: 2px solid var(--user-accent);
  outline-offset: -2px;
}
.accordion-icon {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border-radius: 9px;
  background: var(--user-icon-bg);
  color: var(--user-primary);
  flex: none;
}

.accordion-chevron {
  color: var(--user-text-dim);
  flex: none;
  transition:
    transform 0.2s ease,
    color 0.2s;

  margin-left: auto;
}

.accordion-block.open .accordion-chevron {
  transform: rotate(90deg);
  color: var(--user-accent);
}

.accordion-body {
  padding: 16px 16px 20px;
  border-top: 1px solid var(--user-border);
  animation: accordionOpen 0.18s ease both;
}
@keyframes accordionOpen {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@media (max-width: 720px) {
  .accordion-header {
    padding: 12px;
    gap: 10px;
  }
  .accordion-body {
    padding: 14px 12px 16px;
  }
}
</style>
