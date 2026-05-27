import { RiListCheck3, RiRobot2Line, RiSettings3Line } from "react-icons/ri";
import { NavLink, useLocation } from "react-router-dom";
import { navigationItems } from "../config/controlCenterConfig.js";
import { cn } from "../utils/cn.js";

const navigationIcons = {
  missions: RiListCheck3,
  robots: RiRobot2Line,
  system: RiSettings3Line
};

export default function AppSidebar() {
  const location = useLocation();
  return (
    <aside className="grid h-screen min-w-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden border-r border-slate-500/10 bg-[#080a0e] px-3 py-3 text-slate-100 max-[900px]:h-auto max-[900px]:grid-rows-none max-[900px]:gap-3 max-[900px]:border-b max-[900px]:border-r-0">
      <NavLink
        aria-label="홈으로 이동"
        className="mb-3 rounded-xl border border-slate-500/10 bg-white/[0.03] px-3 py-2 no-underline transition hover:border-sapphire-400/25 hover:bg-sapphire-500/[0.08] max-[900px]:mb-0"
        reloadDocument
        to="/missions"
      >
        <p className="text-[10px] font-black uppercase tracking-normal text-sapphire-300">SST</p>
        <strong className="mt-0.5 block truncate text-sm font-black text-slate-50">Robot Center</strong>
      </NavLink>

      <nav className="grid content-start gap-1.5 overflow-auto max-[900px]:grid-cols-3">
        {navigationItems.map((item) => (
          <NavigationItem
            isSectionActive={location.pathname.startsWith(`${item.path}/`)}
            item={item}
            key={item.key}
          />
        ))}
      </nav>
    </aside>
  );
}

function NavigationItem({ isSectionActive, item }) {
  const Icon = navigationIcons[item.key] ?? RiListCheck3;
  return (
    <NavLink
      className={({ isActive }) => cn(
        "flex h-10 min-w-0 items-center gap-2.5 rounded-lg border border-transparent px-3 text-sm font-extrabold text-slate-400 no-underline transition hover:bg-white/[0.04] hover:text-slate-50",
        (isActive || isSectionActive) && "border-sapphire-300/25 bg-sapphire-500/[0.13] text-white shadow-[inset_3px_0_0_var(--color-sapphire)]"
      )}
      end={item.path !== "/missions"}
      reloadDocument
      to={item.path}
    >
      <Icon aria-hidden className="h-4 w-4 shrink-0 text-current" />
      <span className="min-w-0 truncate">{item.label}</span>
    </NavLink>
  );
}
