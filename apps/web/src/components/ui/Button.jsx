import { cn } from "../../utils/cn.js";

const buttonVariants = {
  danger: "border-red-400/[0.35] bg-red-400/[0.10] text-red-100 hover:border-red-300/60 hover:bg-red-400/[0.16]",
  ghost: "border-transparent bg-transparent text-slate-300 hover:bg-white/[0.06] hover:text-white",
  primary: "border-sapphire-500/35 bg-sapphire-600 text-white hover:bg-sapphire-500",
  secondary: "border-slate-500/30 bg-white/[0.06] text-slate-200 hover:border-sapphire-500/50 hover:bg-sapphire-500/[0.15] hover:text-white"
};

const buttonSizes = {
  icon: "h-8 w-8 p-0",
  md: "h-10 px-4 text-sm",
  sm: "h-8 px-3 text-xs"
};

export default function Button({
  children,
  className,
  size = "md",
  type = "button",
  variant = "secondary",
  ...props
}) {
  return (
    <button
      className={cn(
        "inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-45",
        buttonVariants[variant],
        buttonSizes[size],
        className
      )}
      type={type}
      {...props}
    >
      {children}
    </button>
  );
}
