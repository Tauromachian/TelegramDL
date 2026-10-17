<script setup>
import { useId } from 'vue'

const model = defineModel({ type: String, default: '' })

const props = defineProps({
  label: { type: String, default: '' },
  id: { type: String, default: '' },
  variant: {
    type: String,
    default: '',
    validator(value) {
      return ['', 'default', 'compact'].includes(value)
    }
  }
})

const localId = props.id || useId()
</script>

<template>
  <div class="wrapper">
    <label v-if="label" :for="localId">{{ label }}</label>
    <input
      v-bind="$attrs"
      :id="localId"
      v-model="model"
      class="input"
      :class="{ 'input--compact': variant === 'compact' }"
    />
  </div>
</template>

<style scoped>
.wrapper {
  display: flex;
  flex-direction: column;
}

.label {
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 10px;
  font-weight: 700;
  color: var(--user-text-dim, #94a8bc);
  cursor: pointer;
  background: var(--user-bg-base);
  border: 1px solid var(--user-border-light);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.label--compact {
  font-size: 10px;
  font-weight: 600;
  color: var(--user-text-dim, #94a8bc);
  align-items: center;
  gap: 4px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  height: 14px;
  line-height: 14px;
  margin: 0;
}

.input {
  background: var(--user-bg-base);
  border: 1px solid var(--user-border-light);
  color: var(--user-text-main, #dbe7f5);
  border-radius: 10px;
  padding: 8px 12px;
  outline: none;
  font: inherit;
  font-size: 12px;
  width: 100%;
  transition:
    border-color 0.2s,
    background-color 0.2s,
    color 0.2s;
}

.input--compact {
  height: 32px;
  padding: 0 8px;
  font-size: 11px;
  border-radius: 8px;
}
</style>
