function DocsPage({ t }) {
  const codeClickPixel = `<img src="${API_BASE_URL}/api/click?ad_id=101" style="display:none" />`;
  const codeClickJS = `fetch("${API_BASE_URL}/api/click", { 
        method: "POST", 
        headers: {"Content-Type": "application/json"}, 
        body: JSON.stringify({ ad_id: 101, click_id: "uniq-click-id-123" }) 
        });`;
  const codeConversionCurl = `curl -X POST ${API_BASE_URL}/api/conversion \\
        -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \\
        -H "Content-Type: application/json" \\
        -d '{"ad_id": 101, "order_id": "INV-100500", "revenue": 25.00}'`;
        
  const codeNode = `// Node (fetch)
        // Убедитесь, что у вас есть JWT токен пользователя
        const userToken = "ey..."; 

        await fetch('${API_BASE_URL}/api/conversion', {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json', 
            'Authorization': \`Bearer \${userToken}\`
        },
        body: JSON.stringify({ ad_id: 101, revenue: 25.00, order_id: 'A100' })
        });`;

  return (
    <div className="grid gap-4">
      <Card title={t.TrackClick}>
        <div className="grid md:grid-cols-2 gap-4 text-sm">
          <div>
            <div className="font-semibold mb-1">1) Pixel (GET)</div>
            <pre className="bg-zinc-100 dark:bg-zinc-900 rounded-xl p-3 overflow-auto text-xs whitespace-pre-wrap">{codeClickPixel}</pre>
          </div>
          <div>
            <div className="font-semibold mb-1">2) JavaScript (POST)</div>
            <pre className="bg-zinc-100 dark:bg-zinc-900 rounded-xl p-3 overflow-auto text-xs whitespace-pre-wrap">{codeClickJS}</pre>
          </div>
        </div>
      </Card>
      <Card title={t.SendConversion}>
         <p className="text-sm text-zinc-600 mb-2">
           Эндпоинт <strong>/api/conversion</strong> требует JWT-авторизации. 
           Его нельзя вызывать S2S (сервер-сервер) с API-ключом, его должен вызывать ваш сервер от имени пользователя.
         </p>
        <div className="grid md:grid-cols-2 gap-4 text-xs">
          <pre className="bg-zinc-100 dark:bg-zinc-900 rounded-xl p-3 overflow-auto whitespace-pre-wrap">{codeConversionCurl}</pre>
          <div className="grid gap-3">
            <pre className="bg-zinc-100 dark:bg-zinc-900 rounded-xl p-3 overflow-auto whitespace-pre-wrap">{codeNode}</pre>
          </div>
        </div>
      </Card>
    </div>
  );
}