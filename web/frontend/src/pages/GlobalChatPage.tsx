import { useRouter } from "../contexts/RouterContext";
import { GlobalChatInterface } from "../components/GlobalChatInterface";
import { Button } from "../components/ui/button";
import { ArrowLeft } from "lucide-react";
import { ThemeSwitcher } from "../components/ThemeSwitcher";

export function GlobalChatPage() {
	const { navigate } = useRouter();

	return (
		<div className="text-gray-700 dark:text-gray-100 bg-white dark:bg-gray-900 h-screen max-h-[100dvh] overflow-auto flex flex-col">
			{/* Top Navigation Bar */}
			<div className="h-14 bg-white dark:bg-gray-900 flex items-center px-4 md:px-6 z-10 border-b border-gray-200 dark:border-gray-700">
				<div className="flex items-center gap-3 flex-1">
					{/* Back Button */}
					<Button
						variant="ghost"
						size="sm"
						onClick={() => navigate({ path: "home" })}
						className="gap-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100"
					>
						<ArrowLeft className="h-4 w-4" />
						Back to Home
					</Button>
				</div>
				<ThemeSwitcher />
			</div>

			{/* Main Content Area */}
			<div className="flex-1 overflow-hidden">
				<GlobalChatInterface />
			</div>
		</div>
	);
}
