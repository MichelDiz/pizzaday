<script setup>
import { useWebSocket } from "../composables/useWebSocket";
import InfoCard from "./InfoCard.vue";

const { data } = useWebSocket("ws://localhost:8081/ws");
</script>

<template>
  <div class="p-6 max-w-6xl mx-auto">

    <div v-if="data" class="grid grid-cols-3 gap-4 bg-white shadow-lg p-6 rounded-lg">
      <InfoCard title="Symbol" :value="data.Symbol" />
      <InfoCard title="Last Trade Price" :value="`$${data.LastTradePrice.toFixed(2)}`" />
      <InfoCard title="Total Trades" :value="data.TotalCount" />
      
      <InfoCard title="Buy Count" :value="data.BuyCount" />
      <InfoCard title="Sell Count" :value="data.SellCount" />
      <InfoCard title="Last Update" :value="new Date(data.LastUpdate).toLocaleTimeString()" />
      
      <InfoCard title="Buy Volume" :value="data.BuyVolume.toFixed(4)" />
      <InfoCard title="Sell Volume" :value="data.SellVolume.toFixed(4)" />
      <InfoCard title="Total Frequency" :value="data.TotalFreq.toFixed(2) + ' min⁻¹'" />
      
      <InfoCard title="Buy Activity" :value="data.BuyActivity.toFixed(4)" />
      <InfoCard title="Sell Activity" :value="data.SellActivity.toFixed(4)" />
      <InfoCard title="Kd Ratio" :value="`${data.KdBuy.toFixed(2)} / ${data.KdSell.toFixed(2)}`" />
    </div>

    <p v-else class="text-center text-gray-500 mt-10">Loading data...</p>
  </div>
</template>