import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface MovieCardProps {
  title: string;
  year: number;
  cover?: string;
}

export function MovieCard({ title, year, cover }: MovieCardProps) {
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
      <CardFooter>
        <p className="text-sm text-muted-foreground">
          {year || "Unknown Year"}
        </p>
      </CardFooter>
    </Card>
  );
}
