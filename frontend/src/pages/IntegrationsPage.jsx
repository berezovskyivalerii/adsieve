function IntegrationsPage({ t }) {
  const [googleAccounts, setGoogleAccounts] = useState([]);
  const [selectedAccounts, setSelectedAccounts] = useState(new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchAccounts = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const res = await getGoogleAccounts();
      setGoogleAccounts(res.data?.customer_ids || []);
    } catch (err) {
      console.error(err);
      const errorMsg = err.response?.data?.error || "Failed to fetch accounts";
      setError(errorMsg);
      if (errorMsg.includes("re-consent required")) {
      }
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchAccounts();
  }, [fetchAccounts]);

  const handleConnect = async () => {
    try {
      const res = await connectGoogle();
      const { redirect_url } = res.data;
      if (redirect_url) {
        const authWindow = window.open(redirect_url, "_blank", "width=500,height=600");
        const timer = setInterval(() => {
          if (authWindow.closed) {
            clearInterval(timer);
            fetchAccounts();
          }
        }, 1000);
      }
    } catch (err) {
      console.error(err);
      setError(err.response?.data?.error || "Failed to start connection");
    }
  };

  const handleLink = async () => {
    try {
      await linkGoogleAccounts(Array.from(selectedAccounts));
      alert("Аккаунты успешно привязаны!");
    } catch (err) {
      console.error(err);
      setError(err.response?.data?.error || "Failed to link accounts");
    }
  };

  const toggleAccount = (id) => {
    const newSet = new Set(selectedAccounts);
    if (newSet.has(id)) {
      newSet.delete(id);
    } else {
      newSet.add(id);
    }
    setSelectedAccounts(newSet);
  };

  return (
    <div className="grid gap-4">
      <Card title="Google Ads">
        <div className="flex items-center justify-between">
          <div>
            <div className="font-medium">{t.ConnectGoogle}</div>
            <div className="text-sm text-zinc-500">OAuth → /integrations/google/connect</div>
          </div>
          <button onClick={handleConnect} className="rounded-xl text-sm px-4 py-2 text-white" style={{ backgroundColor: BRAND.navy }}>
            {t.ConnectGoogle}
          </button>
        </div>
        {error && <p className="text-red-500 text-sm mt-3">{error}</p>}
      </Card>

      <Card title={t.GoogleAccountsList}>
        {loading ? (
          <p>Загрузка...</p>
        ) : googleAccounts.length > 0 ? (
          <div className="grid gap-2">
            {googleAccounts.map(id => (
              <label key={id} className="flex items-center gap-2 p-2 border rounded-lg">
                <input 
                  type="checkbox" 
                  checked={selectedAccounts.has(id)} 
                  onChange={() => toggleAccount(id)}
                />
                <span>{id}</span>
              </label>
            ))}
            <button 
              onClick={handleLink} 
              disabled={selectedAccounts.size === 0}
              className="rounded-xl text-sm px-4 py-2 text-white mt-2 disabled:opacity-50" 
              style={{ backgroundColor: BRAND.navy }}
            >
              {t.LinkAccounts}
            </button>
          </div>
        ) : (
          <p className="text-sm text-zinc-500">{t.NoGoogleAccounts}</p>
        )}
      </Card>
      
      <Card title="Server-to-server conversion security">
        <div className="font-medium">{t.APIKey}</div>
      </Card>
    </div>
  );
}