import { computed, onMounted, onUnmounted, ref } from "vue";
import type { Snapshot } from "../types";

const POLL_MS = 15_000;
export const STALE_MS = 60_000;

export function useNodes() {
  const snapshot = ref<Snapshot | null>(null);
  const lastOk = ref<number | null>(null);
  const error = ref<string | null>(null);
  const now = ref(Date.now());
  const startedAt = Date.now();

  async function refresh() {
    try {
      const res = await fetch("/api/nodes", { cache: "no-store" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      snapshot.value = await res.json();
      lastOk.value = Date.now();
      error.value = null;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
    }
  }

  // Before the first success, "stale" means the page has been up for STALE_MS with no data.
  const stale = computed(() => now.value - (lastOk.value ?? startedAt) > STALE_MS);

  let poll: number | undefined;
  let tick: number | undefined;
  onMounted(() => {
    void refresh();
    poll = window.setInterval(refresh, POLL_MS);
    tick = window.setInterval(() => (now.value = Date.now()), 1000);
  });
  onUnmounted(() => {
    window.clearInterval(poll);
    window.clearInterval(tick);
  });

  return { snapshot, lastOk, error, now, stale };
}
