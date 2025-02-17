import { ref, onMounted, onUnmounted } from "vue";

export function useWebSocket(url) {
  const data = ref(null);
  let socket = null;

  onMounted(() => {
    socket = new WebSocket(url);

    socket.onopen = () => console.log("WebSocket Connected");
    socket.onmessage = (event) => {
      try {
        data.value = JSON.parse(event.data);
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };
    socket.onerror = (error) => console.error("WebSocket Error:", error);
    socket.onclose = () => console.log("WebSocket Disconnected");
  });

  onUnmounted(() => {
    if (socket) socket.close();
  });

  return { data };
}