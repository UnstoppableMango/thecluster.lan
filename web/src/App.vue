<script setup>
import { ref } from "vue";

const status = ref("Idle");
const loading = ref(false);

async function ping() {
  loading.value = true;
  status.value = "Pinging...";

  try {
    const response = await fetch("/ping");

    if (!response.ok) {
      throw new Error(`Request failed with ${response.status}`);
    }

    const payload = await response.json();
    status.value = payload.message;
  } catch (error) {
    status.value =
      error instanceof Error ? error.message : "Unable to reach the API";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <main
    class="flex min-h-screen items-center justify-center bg-slate-950 px-6 text-slate-100"
  >
    <section class="w-full max-w-xl rounded-3xl border border-slate-800 bg-slate-900/80 p-10 shadow-2xl shadow-slate-950/50">
      <p class="text-sm font-semibold uppercase tracking-[0.35em] text-cyan-400">
        THECLUSTER
      </p>
      <h1 class="mt-4 text-4xl font-semibold tracking-tight">
        Internal dashboard
      </h1>
      <p class="mt-4 text-base text-slate-300">
        The first slice is intentionally small: a Vue landing page, a Go API,
        and a Nix-native build for both.
      </p>

      <div class="mt-8 flex flex-col gap-4 sm:flex-row sm:items-center">
        <button
          class="inline-flex items-center justify-center rounded-full bg-cyan-400 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-cyan-300 disabled:cursor-not-allowed disabled:bg-cyan-700"
          type="button"
          :disabled="loading"
          @click="ping"
        >
          {{ loading ? "Pinging..." : "Ping API" }}
        </button>

        <p class="text-sm text-slate-300">
          Status:
          <span class="font-medium text-white">{{ status }}</span>
        </p>
      </div>
    </section>
  </main>
</template>
