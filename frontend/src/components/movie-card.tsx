import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { RefreshCw } from "lucide-react";

interface MovieCardProps {
  id: number;
  title: string;
  year: number;
  cover?: string;
  onRefresh?: (id: number) => void;
}

export function MovieCard({
  id,
  title,
  year,
  cover,
  onRefresh,
}: MovieCardProps) {
  return (
    <Card className="overflow-hidden">
      <CardHeader>
        <CardTitle className="line-clamp-1" title={title}>
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {cover ? (
          <img
            src={cover}
            alt={title}
            className="aspect-[2/3] object-cover rounded-md"
          />
        ) : (
          <div className="aspect-[2/3] bg-muted flex items-center justify-center rounded-md">
            No Cover
          </div>
        )}
      </CardContent>
      <CardFooter className="flex justify-between items-center">
        <p className="text-sm text-muted-foreground">
          {year || "Unknown Year"}
        </p>
        {onRefresh && (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onRefresh(id)}
            title="Refresh movie info"
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
