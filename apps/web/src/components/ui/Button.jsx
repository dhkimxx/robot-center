import { cn } from "../../utils/cn.js";

const buttonVariants = {
  danger: "border-red-400/[0.28] bg-red-400/[0.08] text-red-100 hover:border-red-300/50 hover:bg-red-400/[0.14]",
  ghost: "border-transparent bg-transparent text-slate-400 hover:bg-white/[0.05] hover:text-slate-100",
  primary: "border-sapphire-500/40 bg-sapphire-600 text-white shadow-[0_8px_20px_rgba(37,99,235,0.22)] hover:bg-sapphire-500",
  secondary: "border-slate-500/20 bg-white/[0.045] text-slate-200 hover:border-sapphire-400/35 hover:bg-sapphire-500/[0.10] hover:text-white"
};

const buttonSizes = {
  icon: "h-8 w-8 p-0",
  md: "h-9 px-3.5 text-sm",
  sm: "h-8 px-3 text-xs"
};

export default function Button({
  as: Component = "button",
  children,
  className,
  disabled = false,
  size = "md",
  type = "button",
  variant = "secondary",
  ...props
}) {
  const isButton = Component === "button";

  return (
    <Component
      className={cn(
        "inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border font-bold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sapphire-500 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40",
        !isButton && disabled && "pointer-events-none opacity-40",
        buttonVariants[variant],
        buttonSizes[size],
        className
      )}
      disabled={isButton ? disabled : undefined}
      type={isButton ? type : undefined}
      {...props}
    >
      {children}
    </Component>
  );
}
