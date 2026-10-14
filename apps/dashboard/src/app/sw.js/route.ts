export async function GET() {
  const content = `
self.addEventListener('install', function() {
  self.skipWaiting();
});
self.addEventListener('activate', function() {
  self.registration.unregister().then(function() {
    return self.clients.matchAll({ type: 'window' });
  }).then(function(clients) {
    clients.forEach(function(client) {
      client.navigate(client.url);
    });
  });
});
`;
  return new Response(content, {
    headers: {
      "Content-Type": "application/javascript",
      "Cache-Control": "no-store, no-cache, must-revalidate",
    },
  });
}
