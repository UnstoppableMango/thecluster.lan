<script setup lang="ts">
import { computed } from "vue";
import { formatClock, formatPct } from "../format";
import type { Snapshot } from "../types";

const props = defineProps<{ snapshot: Snapshot | null; now: number; lastOk: number | null }>();

const total = computed(() => props.snapshot?.nodes.length ?? 0);
const healthy = computed(() => props.snapshot?.nodes.filter((n) => n.health === "ok").length ?? 0);
const allOk = computed(() => total.value > 0 && healthy.value === total.value);
const age = computed(() => (props.lastOk === null ? null : Math.round((props.now - props.lastOk) / 1000)));
</script>

<template>
  <header class="bar" :class="snapshot === null ? 'is-pending' : allOk ? 'is-ok' : 'is-bad'">
    <span class="brand">THECLUSTER</span>
    <span class="summary">
      <template v-if="snapshot === null">CONNECTING…</template>
      <template v-else-if="allOk">✔ {{ healthy }}/{{ total }} HEALTHY</template>
      <template v-else>✖ {{ total - healthy }} DOWN · {{ healthy }}/{{ total }} HEALTHY</template>
    </span>
    <span class="stats">
      <span>CPU {{ formatPct(snapshot?.cluster.cpuPct ?? null) }}</span>
      <span>MEM {{ formatPct(snapshot?.cluster.memPct ?? null) }}</span>
    </span>
    <span class="clock">
      {{ formatClock(new Date(now)) }}
      <small v-if="age !== null">updated {{ age }}s ago</small>
    </span>
  </header>
</template>

<style scoped>
.bar {
  display: flex;
  align-items: center;
  gap: 2.5rem;
  padding: 0.6rem 1.25rem;
  border-radius: 0.75rem;
  background: var(--surface);
  border: 0.25rem solid var(--health, var(--grid));
  font-variant-numeric: tabular-nums;
}
.is-ok {
  --health: var(--ok);
}
.is-bad {
  --health: var(--bad);
  background: var(--bad);
  color: var(--on-health);
}
.brand {
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: 0.35em;
  color: var(--accent);
}
.is-bad .brand {
  color: inherit;
}
.summary {
  font-size: 2.25rem;
  font-weight: 800;
  letter-spacing: 0.04em;
}
.is-ok .summary {
  color: var(--ok);
}
.stats {
  display: flex;
  gap: 1.75rem;
  font-size: 1.6rem;
  font-weight: 600;
}
.clock {
  margin-left: auto;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.1;
}
.clock small {
  font-size: 0.9rem;
  font-weight: 500;
  opacity: 0.75;
}
</style>
