function SettingsDrawer({ open, onClose, children }) {
  return (
    <div className={classNames("fixed inset-0 z-50 transition", open ? "pointer-events-auto" : "pointer-events-none")}>
      <div className={classNames("absolute inset-0 bg-black/40 transition-opacity", open ? "opacity-100" : "opacity-0")} onClick={onClose} />
      <aside className={classNames("absolute right-0 top-0 h-full w-full sm:w-[420px] bg-white dark:bg-zinc-900 border-l p-5 transition-transform duration-300", open ? "translate-x-0" : "translate-x-full")}>
        <div className="flex items-center justify-between mb-3">
          <div className="text-lg font-semibold" style={{ color: BRAND.navy }}>Settings</div>
          <button className="rounded-xl px-3 py-2 border" style={{ borderColor: BRAND.steel }} onClick={onClose}>Close</button>
        </div>
        {children}
      </aside>
    </div>
  );
}