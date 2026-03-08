import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const navItems = [
    { to: "/catalog", label: "Catalog", roles: null },
    { to: "/inventory", label: "Inventory", roles: ["PurchaseOfficer"] },
    { to: "/admin", label: "Users", roles: ["Admin"] },
];

export default function Navbar() {
    const { user, logout } = useAuth();
    const location = useLocation();
    const navigate = useNavigate();

    if (!user) return null;

    const handleLogout = async () => {
        await logout();
        navigate("/login");
    };

    return (
        <header className="sticky top-0 z-40 border-b border-border/50 bg-background/80 backdrop-blur-lg">
            <div className="mx-auto flex h-14 max-w-5xl items-center px-6">
                <Link to="/" className="mr-8 flex items-center gap-2 font-bold text-lg tracking-tight">

                    <span className="hidden sm:inline">Core Procurement Service</span>
                </Link>

                <nav className="flex items-center gap-0.5">
                    {navItems
                        .filter((item) => !item.roles || item.roles.includes(user.role))
                        .map((item) => {
                            const isActive = location.pathname === item.to;
                            return (
                                <Link key={item.to} to={item.to}>
                                    <button
                                        className={`relative px-3 py-1.5 text-sm font-medium rounded-lg transition-colors duration-150 ${isActive
                                                ? "text-foreground"
                                                : "text-muted-foreground hover:text-foreground"
                                            }`}
                                    >
                                        {item.label}
                                        {isActive && (
                                            <span className="absolute inset-x-1 -bottom-[13px] h-0.5 rounded-full bg-primary" />
                                        )}
                                    </button>
                                </Link>
                            );
                        })}
                </nav>

                <div className="ml-auto">
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="gap-2 text-muted-foreground hover:text-foreground"
                            >
                                <span className="material-symbols-outlined text-[20px]">account_circle</span>
                                <span className="hidden sm:inline">{user.first_name} {user.last_name}</span>
                                <span className="inline sm:hidden">{user.first_name?.[0]}{user.last_name?.[0]}</span>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end" className="w-48">
                            <div className="px-2 py-1.5">
                                <p className="text-sm font-medium">{user.first_name} {user.last_name}</p>
                                <p className="text-xs text-muted-foreground">{user.role}</p>
                            </div>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => navigate("/profile")} className="cursor-pointer">
                                Profile
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={handleLogout} className="cursor-pointer text-destructive focus:text-destructive">
                                Logout
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </div>
        </header>
    );
}
