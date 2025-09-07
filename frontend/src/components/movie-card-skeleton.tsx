import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

export function MovieCardSkeleton() {
  return (
    <Card className="overflow-hidden">
      <CardHeader>
        <Skeleton className="h-6 w-[80%]" />
      </CardHeader>
      <CardContent>
        <Skeleton className="aspect-[2/3] w-full" />
      </CardContent>
      <CardFooter>
        <Skeleton className="h-4 w-[40%]" />
      </CardFooter>
    </Card>
  );
}
