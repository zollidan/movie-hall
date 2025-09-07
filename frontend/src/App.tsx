import { ModeToggle } from "@/components/mode-toggle";
import { MovieGrid } from "@/components/movie-grid";

function App() {
  return (
    <div className="min-h-screen bg-background">
      <div className="p-4">
        <div className="flex justify-end mb-4">
          <ModeToggle />
        </div>
        <div className="container mx-auto">
          <MovieGrid />
        </div>
      </div>
    </div>
  );
}

export default App;
