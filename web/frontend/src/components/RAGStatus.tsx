import { useState, useEffect } from "react";
import { Database, RefreshCw, CheckCircle, XCircle, AlertCircle } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { useAuth } from "../contexts/AuthContext";
import { useToast } from "./ui/toast";

interface RAGStats {
	status: string;
	transcript_count: number;
	collection_name?: string;
	message?: string;
}

export function RAGStatus() {
	const { getAuthHeaders } = useAuth();
	const { toast } = useToast();
	const [stats, setStats] = useState<RAGStats | null>(null);
	const [loading, setLoading] = useState(true);
	const [refreshing, setRefreshing] = useState(false);

	const fetchStats = async () => {
		try {
			const response = await fetch("/api/v1/rag/stats", {
				headers: getAuthHeaders(),
			});

			if (response.ok) {
				const data = await response.json();
				setStats(data);
			} else {
				const errorData = await response.json();
				setStats({
					status: "error",
					transcript_count: 0,
					message: errorData.error || "Failed to fetch RAG stats",
				});
			}
		} catch (error: any) {
			setStats({
				status: "error",
				transcript_count: 0,
				message: error.message || "Failed to fetch RAG stats",
			});
		} finally {
			setLoading(false);
			setRefreshing(false);
		}
	};

	useEffect(() => {
		fetchStats();
	}, []);

	const handleRefresh = async () => {
		setRefreshing(true);
		await fetchStats();
		toast({
			title: "Refreshed",
			description: "RAG statistics updated",
		});
	};

	const handleBackfill = async () => {
		setRefreshing(true);
		try {
			const response = await fetch("/api/v1/rag/backfill", {
				method: "POST",
				headers: getAuthHeaders(),
			});

			if (response.ok) {
				const data = await response.json();
				toast({
					title: "Backfill Complete",
					description: `Processed ${data.processed} transcriptions, ${data.failed} failed`,
				});
				await fetchStats();
			} else {
				const errorData = await response.json();
				toast({
					title: "Error",
					description: errorData.error || "Failed to backfill",
				});
			}
		} catch (error: any) {
			toast({
				title: "Error",
				description: error.message || "Failed to backfill",
			});
		} finally {
			setRefreshing(false);
		}
	};

	if (loading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Database className="h-5 w-5" />
						RAG Status
					</CardTitle>
					<CardDescription>Retrieval-Augmented Generation system status</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="flex items-center justify-center py-8">
						<div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
					</div>
				</CardContent>
			</Card>
		);
	}

	const isActive = stats?.status === "active";
	const transcriptCount = stats?.transcript_count || 0;

	return (
		<Card>
			<CardHeader>
				<div className="flex items-center justify-between">
					<div>
						<CardTitle className="flex items-center gap-2">
							<Database className="h-5 w-5" />
							RAG Status
						</CardTitle>
						<CardDescription>Retrieval-Augmented Generation system status</CardDescription>
					</div>
					<Button
						variant="outline"
						size="sm"
						onClick={handleRefresh}
						disabled={refreshing}
						className="gap-2"
					>
						<RefreshCw className={`h-4 w-4 ${refreshing ? "animate-spin" : ""}`} />
						Refresh
					</Button>
				</div>
			</CardHeader>
			<CardContent className="space-y-4">
				{/* Status Indicator */}
				<div className="flex items-center gap-3">
					{isActive ? (
						<CheckCircle className="h-5 w-5 text-green-500" />
					) : stats?.status === "error" ? (
						<XCircle className="h-5 w-5 text-red-500" />
					) : (
						<AlertCircle className="h-5 w-5 text-yellow-500" />
					)}
					<div>
						<div className="font-medium text-gray-900 dark:text-gray-100">
							Status: {isActive ? "Active" : stats?.status || "Unknown"}
						</div>
						{stats?.message && (
							<div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
								{stats.message}
							</div>
						)}
					</div>
				</div>

				{/* Transcript Count */}
				{isActive && (
					<div className="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
						<div className="flex items-center justify-between">
							<div>
								<div className="text-sm text-gray-600 dark:text-gray-400">
									Transcripts in RAG
								</div>
								<div className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">
									{transcriptCount}
								</div>
							</div>
							<Database className="h-8 w-8 text-blue-500 opacity-50" />
						</div>
					</div>
				)}

				{/* Collection Info */}
				{isActive && stats?.collection_name && (
					<div className="text-sm text-gray-600 dark:text-gray-400">
						Collection: <span className="font-mono text-gray-900 dark:text-gray-100">{stats.collection_name}</span>
					</div>
				)}

				{/* Backfill Button */}
				{isActive && (
					<Button
						variant="outline"
						onClick={handleBackfill}
						disabled={refreshing}
						className="w-full"
					>
						{refreshing ? (
							<>
								<RefreshCw className="h-4 w-4 mr-2 animate-spin" />
								Processing...
							</>
						) : (
							<>
								<RefreshCw className="h-4 w-4 mr-2" />
								Backfill Existing Transcriptions
							</>
						)}
					</Button>
				)}

				{/* Info Message */}
				{isActive && transcriptCount === 0 && (
					<div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3 text-sm text-blue-700 dark:text-blue-400">
						No transcripts in RAG yet. Upload and transcribe audio files to populate the RAG system.
					</div>
				)}
			</CardContent>
		</Card>
	);
}
