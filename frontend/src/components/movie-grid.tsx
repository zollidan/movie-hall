import { useEffect, useState } from "react";
import type { Movie } from "@/types/movie";
import { MovieCard } from "./movie-card";
import { MovieCardSkeleton } from "./movie-card-skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { Toast } from "@/components/ui/toast";

export function MovieGrid() {
  const [movies, setMovies] = useState<Movie[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState<number | null>(null);
  const [toast, setToast] = useState<{
    type: "success" | "error";
    title: string;
    description?: string;
  } | null>(null);

  useEffect(() => {
    const fetchMovies = async () => {
      try {
        const response = await fetch("http://localhost:8080/api/library");
        if (!response.ok) {
          throw new Error("Failed to fetch movies");
        }
        const data = await response.json();
        setMovies(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch movies");
      } finally {
        setIsLoading(false);
      }
    };

    fetchMovies();
  }, []);

  if (isLoading) {
    return (
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
        {Array.from({ length: 10 }).map((_, index) => (
          <MovieCardSkeleton key={index} />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-2xl mx-auto mt-4">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    );
  }

  if (movies.length === 0) {
    return (
      <div className="max-w-2xl mx-auto mt-4">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>No Movies Found</AlertTitle>
          <AlertDescription>
            There are no movies in your library. Try adding some movies or
            checking your library path.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const handleRefresh = async (id: number) => {
    try {
      setRefreshing(id);
      const response = await fetch(
        `http://localhost:8080/api/movies/${id}/refresh`,
        {
          method: "POST",
        }
      );

      if (!response.ok) {
        throw new Error("Failed to refresh movie info");
      }

      const updatedMovie = await response.json();
      setMovies(
        movies.map((movie) => (movie.ID === id ? updatedMovie : movie))
      );

      setToast({
        type: "success",
        title: "Movie Updated",
        description: "Movie information has been refreshed",
      });
    } catch (err) {
      setToast({
        type: "error",
        title: "Update Failed",
        description:
          err instanceof Error ? err.message : "Failed to update movie info",
      });
    } finally {
      setRefreshing(null);
      setTimeout(() => setToast(null), 3000);
    }
  };

  return (
    <>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
        {movies.map((movie) => (
          <MovieCard
            key={movie.ID}
            id={movie.ID}
            title={movie.Title}
            year={movie.Year}
            cover={movie.Cover}
            onRefresh={handleRefresh}
          />
        ))}
      </div>
      {toast && <Toast {...toast} />}
    </>
  );
}
